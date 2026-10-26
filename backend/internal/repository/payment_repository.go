package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"kopirise/backend/internal/database"
	"kopirise/backend/internal/models"
)

var (
	ErrPaymentNotFound      = errors.New("transaksi pembayaran tidak ditemukan")
	ErrOverpayment          = errors.New("jumlah pembayaran melebihi sisa tagihan pesanan")
	ErrOrderAlreadyPaid     = errors.New("pesanan sudah lunas sepenuhnya")
	ErrNoActiveShift        = errors.New("kasir belum membuka shift kas aktif")
	ErrActiveShiftExists    = errors.New("kasir sudah memiliki shift kas aktif yang belum ditutup")
	ErrShiftAlreadyClosed   = errors.New("shift kas ini sudah ditutup")
	ErrDuplicatePaymentIdem = errors.New("pembayaran dengan idempotency key ini sudah pernah diproses")
	ErrRefundExceedsPaid    = errors.New("jumlah refund melebihi total nominal pembayaran yang sah")
)

type PaymentRepository struct {
	db *database.DB
}

func NewPaymentRepository(db *database.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// CreatePayment mencatat pembayaran baru, memperbarui saldo akun keuangan, dan mengupdate status pesanan secara atomik.
func (r *PaymentRepository) CreatePayment(ctx context.Context, p *models.Payment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// 1. Ambil order dan lock / hitung sisa tagihan
	var totalAmount, paidAmount int64
	var orderStatus string
	queryOrder := `SELECT total_amount, paid_amount, status FROM orders WHERE id = ?;`
	if err := tx.QueryRowContext(ctx, queryOrder, p.OrderID).Scan(&totalAmount, &paidAmount, &orderStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrOrderNotFound
		}
		return err
	}

	remaining := totalAmount - paidAmount
	if remaining <= 0 {
		return ErrOrderAlreadyPaid
	}
	if p.Amount > remaining {
		return ErrOverpayment
	}

	// 2. Insert payment record
	insertPaySQL := `
		INSERT INTO payments (
			id, order_id, cash_shift_id, payment_number, idempotency_key,
			payment_method, amount, cash_tendered, change_due, reference_number,
			processed_by_user_id, financial_account_id, status, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertPaySQL,
		p.ID,
		p.OrderID,
		p.CashShiftID,
		p.PaymentNumber,
		p.IdempotencyKey,
		p.PaymentMethod,
		p.Amount,
		p.CashTendered,
		p.ChangeDue,
		p.ReferenceNumber,
		p.ProcessedByUserID,
		p.FinancialAccountID,
		p.Status,
		p.Notes,
		nowStr,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: payments.idempotency_key") {
			return ErrDuplicatePaymentIdem
		}
		return fmt.Errorf("insert payment gagal: %w", err)
	}

	// 3. Update Order paid amount & status
	newPaidAmount := paidAmount + p.Amount
	newPaymentStatus := "partially_paid"
	if newPaidAmount >= totalAmount {
		newPaymentStatus = "paid"
	}

	// Update order status: jika sebelumnya pending atau confirmed, ubah ke preparing/confirmed
	newOrderStatus := orderStatus
	if orderStatus == "pending" {
		newOrderStatus = "confirmed"
	}

	updateOrderSQL := `
		UPDATE orders
		SET paid_amount = ?, payment_status = ?, status = ?, updated_at = ?
		WHERE id = ?;
	`
	if _, err := tx.ExecContext(ctx, updateOrderSQL, newPaidAmount, newPaymentStatus, newOrderStatus, nowStr, p.OrderID); err != nil {
		return fmt.Errorf("update order payment status gagal: %w", err)
	}

	// 4. Update saldo Financial Account (+ amount)
	updateAccSQL := `UPDATE financial_accounts SET current_balance = current_balance + ?, updated_at = ? WHERE id = ?;`
	if _, err := tx.ExecContext(ctx, updateAccSQL, p.Amount, nowStr, p.FinancialAccountID); err != nil {
		return fmt.Errorf("update saldo akun keuangan gagal: %w", err)
	}

	// 5. Jika metode tunai dan terhubung dengan shift kas, update cash_sales & expected_cash
	if p.PaymentMethod == "cash" && p.CashShiftID != nil && *p.CashShiftID != "" {
		updateShiftSQL := `
			UPDATE cash_shifts
			SET cash_sales = cash_sales + ?, expected_cash = expected_cash + ?, updated_at = ?
			WHERE id = ?;
		`
		if _, err := tx.ExecContext(ctx, updateShiftSQL, p.Amount, p.Amount, nowStr, *p.CashShiftID); err != nil {
			return fmt.Errorf("update kas shift gagal: %w", err)
		}
	}

	return tx.Commit()
}

// CreateRefund mencatat pengembalian dana, menyesuaikan saldo akun keuangan, dan mencatat histori tanpa menghapus data asli.
func (r *PaymentRepository) CreateRefund(ctx context.Context, ref *models.Refund) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// Ambil payment asli
	var payAmount int64
	var payStatus string
	queryPay := `SELECT amount, status FROM payments WHERE id = ?;`
	if err := tx.QueryRowContext(ctx, queryPay, ref.PaymentID).Scan(&payAmount, &payStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPaymentNotFound
		}
		return err
	}

	if ref.Amount > payAmount {
		return ErrRefundExceedsPaid
	}

	// Insert refund record
	insertRefSQL := `
		INSERT INTO refunds (
			id, refund_number, payment_id, order_id, cash_shift_id,
			amount, reason, refund_method, financial_account_id,
			approved_by_user_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertRefSQL,
		ref.ID,
		ref.RefundNumber,
		ref.PaymentID,
		ref.OrderID,
		ref.CashShiftID,
		ref.Amount,
		ref.Reason,
		ref.RefundMethod,
		ref.FinancialAccountID,
		ref.ApprovedByUserID,
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("insert refund gagal: %w", err)
	}

	// Update order
	var totalAmount, paidAmount int64
	queryOrder := `SELECT total_amount, paid_amount FROM orders WHERE id = ?;`
	_ = tx.QueryRowContext(ctx, queryOrder, ref.OrderID).Scan(&totalAmount, &paidAmount)

	newPaid := paidAmount - ref.Amount
	if newPaid < 0 {
		newPaid = 0
	}
	newPaymentStatus := "partially_refunded"
	if newPaid == 0 {
		newPaymentStatus = "refunded"
	}

	updateOrderSQL := `UPDATE orders SET paid_amount = ?, payment_status = ?, updated_at = ? WHERE id = ?;`
	_, _ = tx.ExecContext(ctx, updateOrderSQL, newPaid, newPaymentStatus, nowStr, ref.OrderID)

	// Potong saldo financial account
	updateAccSQL := `UPDATE financial_accounts SET current_balance = current_balance - ?, updated_at = ? WHERE id = ?;`
	_, _ = tx.ExecContext(ctx, updateAccSQL, ref.Amount, nowStr, ref.FinancialAccountID)

	// Jika refund tunai dan ada shift kas, update cash_refunds & expected_cash
	if ref.RefundMethod == "cash" && ref.CashShiftID != nil && *ref.CashShiftID != "" {
		updateShiftSQL := `
			UPDATE cash_shifts
			SET cash_refunds = cash_refunds + ?, expected_cash = expected_cash - ?, updated_at = ?
			WHERE id = ?;
		`
		_, _ = tx.ExecContext(ctx, updateShiftSQL, ref.Amount, ref.Amount, nowStr, *ref.CashShiftID)
	}

	return tx.Commit()
}

// Cash Shift Management

func (r *PaymentRepository) GetActiveCashShift(ctx context.Context, cashierUserID string) (*models.CashShift, error) {
	query := `
		SELECT s.id, s.cashier_user_id, u.name, s.shift_number, s.opened_at, s.initial_cash,
		       s.cash_sales, s.cash_in, s.cash_out, s.cash_refunds, s.expected_cash, s.status,
		       s.created_at, s.updated_at
		FROM cash_shifts s
		JOIN users u ON s.cashier_user_id = u.id
		WHERE s.cashier_user_id = ? AND s.status = 'open'
		ORDER BY s.opened_at DESC
		LIMIT 1;
	`
	shf := &models.CashShift{}
	var openedStr, createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, cashierUserID).Scan(
		&shf.ID, &shf.CashierUserID, &shf.CashierName, &shf.ShiftNumber,
		&openedStr, &shf.InitialCash, &shf.CashSales, &shf.CashIn, &shf.CashOut,
		&shf.CashRefunds, &shf.ExpectedCash, &shf.Status, &createdStr, &updatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActiveShift
		}
		return nil, err
	}

	shf.OpenedAt, _ = time.Parse(time.RFC3339, openedStr)
	shf.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	shf.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	// Fetch movements
	shf.Movements, _ = r.getCashMovements(ctx, shf.ID)

	return shf, nil
}

func (r *PaymentRepository) OpenCashShift(ctx context.Context, s *models.CashShift) error {
	// Pastikan tidak ada shift open untuk kasir ini
	var count int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cash_shifts WHERE cashier_user_id = ? AND status = 'open';", s.CashierUserID).Scan(&count)
	if count > 0 {
		return ErrActiveShiftExists
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO cash_shifts (
			id, cashier_user_id, shift_number, opened_at, initial_cash,
			cash_sales, cash_in, cash_out, cash_refunds, expected_cash,
			status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, 0, 0, 0, 0, ?, 'open', ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		s.ID,
		s.CashierUserID,
		s.ShiftNumber,
		nowStr,
		s.InitialCash,
		s.InitialCash, // Expected cash awal = modal awal
		nowStr,
		nowStr,
	)
	return err
}

func (r *PaymentRepository) CloseCashShift(ctx context.Context, shiftID string, actualCash, discrepancy int64, notes string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `
		UPDATE cash_shifts
		SET status = 'closed', closed_at = ?, actual_cash = ?, discrepancy = ?, closing_notes = ?, updated_at = ?
		WHERE id = ? AND status = 'open';
	`
	res, err := r.db.ExecContext(ctx, query, nowStr, actualCash, discrepancy, notes, nowStr, shiftID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrShiftAlreadyClosed
	}
	return nil
}

func (r *PaymentRepository) CreateCashMovement(ctx context.Context, m *models.CashMovement) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	insertSQL := `
		INSERT INTO cash_movements (id, cash_shift_id, type, amount, reason, authorized_by_user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	if _, err := tx.ExecContext(ctx, insertSQL, m.ID, m.CashShiftID, m.Type, m.Amount, m.Reason, m.AuthorizedByUserID, nowStr); err != nil {
		return err
	}

	// Update shift totals
	var updateShiftSQL string
	if m.Type == "cash_in" {
		updateShiftSQL = "UPDATE cash_shifts SET cash_in = cash_in + ?, expected_cash = expected_cash + ?, updated_at = ? WHERE id = ?;"
	} else {
		updateShiftSQL = "UPDATE cash_shifts SET cash_out = cash_out + ?, expected_cash = expected_cash - ?, updated_at = ? WHERE id = ?;"
	}

	if _, err := tx.ExecContext(ctx, updateShiftSQL, m.Amount, m.Amount, nowStr, m.CashShiftID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PaymentRepository) getCashMovements(ctx context.Context, shiftID string) ([]*models.CashMovement, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, cash_shift_id, type, amount, reason, authorized_by_user_id, created_at FROM cash_movements WHERE cash_shift_id = ? ORDER BY created_at ASC;", shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.CashMovement
	for rows.Next() {
		m := &models.CashMovement{}
		var authID sql.NullString
		var createdStr string
		if err := rows.Scan(&m.ID, &m.CashShiftID, &m.Type, &m.Amount, &m.Reason, &authID, &createdStr); err != nil {
			return nil, err
		}
		if authID.Valid {
			m.AuthorizedByUserID = &authID.String
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		list = append(list, m)
	}
	return list, nil
}
