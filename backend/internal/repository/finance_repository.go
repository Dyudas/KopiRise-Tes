package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"kopirise/backend/internal/models"
)

var (
	ErrAccountNotFound     = errors.New("akun keuangan tidak ditemukan")
	ErrInsufficientBalance = errors.New("saldo akun sumber tidak mencukupi untuk transfer")
	ErrSameAccountTransfer = errors.New("akun sumber dan tujuan transfer tidak boleh sama")
	ErrExpenseNotFound     = errors.New("biaya pengeluaran tidak ditemukan")
)

type FinanceRepository struct {
	db *sql.DB
}

func NewFinanceRepository(db *sql.DB) *FinanceRepository {
	return &FinanceRepository{db: db}
}

// ==========================================
// 1. BAGAN AKUN KEUANGAN (FINANCIAL ACCOUNTS)
// ==========================================

func (r *FinanceRepository) ListAccounts(ctx context.Context) ([]*models.FinancialAccount, error) {
	query := `
		SELECT id, name, account_type, COALESCE(account_number, ''), current_balance, is_active, created_at, updated_at
		FROM financial_accounts
		ORDER BY account_type ASC, name ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list accounts error: %w", err)
	}
	defer rows.Close()

	var accounts []*models.FinancialAccount
	for rows.Next() {
		var a models.FinancialAccount
		var isActiveInt int
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&a.ID, &a.Name, &a.AccountType, &a.AccountNumber,
			&a.CurrentBalance, &isActiveInt, &createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, err
		}
		a.IsActive = isActiveInt == 1
		a.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		accounts = append(accounts, &a)
	}
	return accounts, nil
}

func (r *FinanceRepository) GetAccountByID(ctx context.Context, id string) (*models.FinancialAccount, error) {
	query := `
		SELECT id, name, account_type, COALESCE(account_number, ''), current_balance, is_active, created_at, updated_at
		FROM financial_accounts
		WHERE id = ?;
	`
	var a models.FinancialAccount
	var isActiveInt int
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Name, &a.AccountType, &a.AccountNumber,
		&a.CurrentBalance, &isActiveInt, &createdAtStr, &updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	a.IsActive = isActiveInt == 1
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &a, nil
}

// CreateTransfer mutasi dana antar rekening secara atomik dalam satu SQL Transaction.
func (r *FinanceRepository) CreateTransfer(ctx context.Context, t *models.AccountTransfer) error {
	if t.FromAccountID == t.ToAccountID {
		return ErrSameAccountTransfer
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Periksa saldo akun sumber
	var fromBalance int64
	err = tx.QueryRowContext(ctx, "SELECT current_balance FROM financial_accounts WHERE id = ?;", t.FromAccountID).Scan(&fromBalance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountNotFound
		}
		return err
	}
	if fromBalance < t.Amount {
		return ErrInsufficientBalance
	}

	// 2. Kurangi akun sumber
	nowStr := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.ExecContext(ctx, "UPDATE financial_accounts SET current_balance = current_balance - ?, updated_at = ? WHERE id = ?;", t.Amount, nowStr, t.FromAccountID)
	if err != nil {
		return fmt.Errorf("potong saldo sumber gagal: %w", err)
	}

	// 3. Tambahkan akun tujuan
	res, err := tx.ExecContext(ctx, "UPDATE financial_accounts SET current_balance = current_balance + ?, updated_at = ? WHERE id = ?;", t.Amount, nowStr, t.ToAccountID)
	if err != nil {
		return fmt.Errorf("tambah saldo tujuan gagal: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrAccountNotFound
	}

	// 4. Catat riwayat transfer
	insertSQL := `
		INSERT INTO account_transfers (
			id, transfer_number, from_account_id, to_account_id, amount, transfer_date, description, created_by_user_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertSQL,
		t.ID, t.TransferNumber, t.FromAccountID, t.ToAccountID, t.Amount, t.TransferDate, t.Description, t.CreatedByUserID, nowStr,
	)
	if err != nil {
		return fmt.Errorf("insert transfer history gagal: %w", err)
	}

	return tx.Commit()
}

func (r *FinanceRepository) ListTransfers(ctx context.Context, page, perPage int) ([]*models.AccountTransfer, *models.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM account_transfers;").Scan(&total)

	query := `
		SELECT t.id, t.transfer_number, t.from_account_id, fa_from.name, t.to_account_id, fa_to.name,
		       t.amount, t.transfer_date, COALESCE(t.description, ''), t.created_by_user_id, u.name, t.created_at
		FROM account_transfers t
		JOIN financial_accounts fa_from ON t.from_account_id = fa_from.id
		JOIN financial_accounts fa_to ON t.to_account_id = fa_to.id
		JOIN users u ON t.created_by_user_id = u.id
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?;
	`
	rows, err := r.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var transfers []*models.AccountTransfer
	for rows.Next() {
		var tr models.AccountTransfer
		var createdAtStr string
		if err := rows.Scan(
			&tr.ID, &tr.TransferNumber, &tr.FromAccountID, &tr.FromAccountName,
			&tr.ToAccountID, &tr.ToAccountName, &tr.Amount, &tr.TransferDate,
			&tr.Description, &tr.CreatedByUserID, &tr.CreatedByUserName, &createdAtStr,
		); err != nil {
			return nil, nil, err
		}
		tr.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		transfers = append(transfers, &tr)
	}

	totalPages := (total + perPage - 1) / perPage
	meta := &models.PaginationMeta{
		CurrentPage: page,
		PerPage:     perPage,
		TotalItems:  total,
		TotalPages:  totalPages,
	}
	return transfers, meta, nil
}

// ==========================================
// 2. BIAYA PENGELUARAN (EXPENSES)
// ==========================================

func (r *FinanceRepository) ListExpenseCategories(ctx context.Context) ([]*models.ExpenseCategory, error) {
	query := `SELECT id, name, COALESCE(description, ''), created_at FROM expense_categories ORDER BY name ASC;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []*models.ExpenseCategory
	for rows.Next() {
		var c models.ExpenseCategory
		var createdAtStr string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &createdAtStr); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		cats = append(cats, &c)
	}
	return cats, nil
}

func (r *FinanceRepository) CreateExpense(ctx context.Context, exp *models.Expense, attachments []*models.ExpenseAttachment) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// 1. Insert Expense Record
	insertSQL := `
		INSERT INTO expenses (
			id, expense_number, expense_date, category_id, financial_account_id,
			amount, description, vendor, reference_number, status,
			created_by_user_id, approved_by_user_id, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertSQL,
		exp.ID, exp.ExpenseNumber, exp.ExpenseDate, exp.CategoryID, exp.FinancialAccountID,
		exp.Amount, exp.Description, exp.Vendor, exp.ReferenceNumber, exp.Status,
		exp.CreatedByUserID, exp.ApprovedByUserID, exp.Notes, nowStr, nowStr,
	)
	if err != nil {
		return fmt.Errorf("insert expense gagal: %w", err)
	}

	// 2. Insert Attachments jika ada
	if len(attachments) > 0 {
		attSQL := `
			INSERT INTO expense_attachments (id, expense_id, file_path, file_name, file_size, mime_type, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?);
		`
		for _, att := range attachments {
			_, err = tx.ExecContext(ctx, attSQL, att.ID, exp.ID, att.FilePath, att.FileName, att.FileSize, att.MimeType, nowStr)
			if err != nil {
				return fmt.Errorf("insert expense attachment gagal: %w", err)
			}
		}
	}

	// 3. Jika status langsung 'approved', kurangi saldo akun finansial secara atomik
	if exp.Status == "approved" {
		_, err = tx.ExecContext(ctx, "UPDATE financial_accounts SET current_balance = current_balance - ?, updated_at = ? WHERE id = ?;", exp.Amount, nowStr, exp.FinancialAccountID)
		if err != nil {
			return fmt.Errorf("potong saldo akun pengeluaran gagal: %w", err)
		}
	}

	return tx.Commit()
}

func (r *FinanceRepository) ListExpenses(ctx context.Context, status, categoryID, startDate, endDate string, page, perPage int) ([]*models.Expense, *models.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var whereClauses []string
	var args []interface{}

	if status != "" {
		whereClauses = append(whereClauses, "e.status = ?")
		args = append(args, status)
	}
	if categoryID != "" {
		whereClauses = append(whereClauses, "e.category_id = ?")
		args = append(args, categoryID)
	}
	if startDate != "" {
		whereClauses = append(whereClauses, "e.expense_date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		whereClauses = append(whereClauses, "e.expense_date <= ?")
		args = append(args, endDate)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countSQL := "SELECT COUNT(*) FROM expenses e " + whereSQL
	var total int
	_ = r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT e.id, e.expense_number, e.expense_date, e.category_id, ec.name,
		       e.financial_account_id, fa.name, e.amount, e.description,
		       COALESCE(e.vendor, ''), COALESCE(e.reference_number, ''), e.status,
		       e.created_by_user_id, u.name, e.approved_by_user_id, COALESCE(ua.name, ''),
		       COALESCE(e.notes, ''), e.created_at, e.updated_at
		FROM expenses e
		JOIN expense_categories ec ON e.category_id = ec.id
		JOIN financial_accounts fa ON e.financial_account_id = fa.id
		JOIN users u ON e.created_by_user_id = u.id
		LEFT JOIN users ua ON e.approved_by_user_id = ua.id
		%s
		ORDER BY e.expense_date DESC, e.created_at DESC
		LIMIT ? OFFSET ?;
	`, whereSQL)

	args = append(args, perPage, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var expenses []*models.Expense
	for rows.Next() {
		var exp models.Expense
		var createdAtStr, updatedAtStr string
		var appID, appName sql.NullString

		if err := rows.Scan(
			&exp.ID, &exp.ExpenseNumber, &exp.ExpenseDate, &exp.CategoryID, &exp.CategoryName,
			&exp.FinancialAccountID, &exp.AccountName, &exp.Amount, &exp.Description,
			&exp.Vendor, &exp.ReferenceNumber, &exp.Status,
			&exp.CreatedByUserID, &exp.CreatedByUserName, &appID, &appName,
			&exp.Notes, &createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, nil, err
		}

		if appID.Valid {
			exp.ApprovedByUserID = &appID.String
			exp.ApprovedByUserName = &appName.String
		}
		exp.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		exp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		expenses = append(expenses, &exp)
	}

	totalPages := (total + perPage - 1) / perPage
	meta := &models.PaginationMeta{
		CurrentPage: page,
		PerPage:     perPage,
		TotalItems:  total,
		TotalPages:  totalPages,
	}
	return expenses, meta, nil
}

func (r *FinanceRepository) GetExpenseByID(ctx context.Context, id string) (*models.Expense, error) {
	query := `
		SELECT e.id, e.expense_number, e.expense_date, e.category_id, ec.name,
		       e.financial_account_id, fa.name, e.amount, e.description,
		       COALESCE(e.vendor, ''), COALESCE(e.reference_number, ''), e.status,
		       e.created_by_user_id, u.name, e.approved_by_user_id, COALESCE(ua.name, ''),
		       COALESCE(e.notes, ''), e.created_at, e.updated_at
		FROM expenses e
		JOIN expense_categories ec ON e.category_id = ec.id
		JOIN financial_accounts fa ON e.financial_account_id = fa.id
		JOIN users u ON e.created_by_user_id = u.id
		LEFT JOIN users ua ON e.approved_by_user_id = ua.id
		WHERE e.id = ?;
	`
	var exp models.Expense
	var createdAtStr, updatedAtStr string
	var appID, appName sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&exp.ID, &exp.ExpenseNumber, &exp.ExpenseDate, &exp.CategoryID, &exp.CategoryName,
		&exp.FinancialAccountID, &exp.AccountName, &exp.Amount, &exp.Description,
		&exp.Vendor, &exp.ReferenceNumber, &exp.Status,
		&exp.CreatedByUserID, &exp.CreatedByUserName, &appID, &appName,
		&exp.Notes, &createdAtStr, &updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}

	if appID.Valid {
		exp.ApprovedByUserID = &appID.String
		exp.ApprovedByUserName = &appName.String
	}
	exp.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	exp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	// Attachments
	attRows, err := r.db.QueryContext(ctx, "SELECT id, file_path, file_name, file_size, mime_type, created_at FROM expense_attachments WHERE expense_id = ?;", id)
	if err == nil {
		defer attRows.Close()
		for attRows.Next() {
			var att models.ExpenseAttachment
			att.ExpenseID = id
			var attCreatedStr string
			if err := attRows.Scan(&att.ID, &att.FilePath, &att.FileName, &att.FileSize, &att.MimeType, &attCreatedStr); err == nil {
				att.CreatedAt, _ = time.Parse(time.RFC3339, attCreatedStr)
				exp.Attachments = append(exp.Attachments, &att)
			}
		}
	}

	return &exp, nil
}

func (r *FinanceRepository) UpdateExpenseStatus(ctx context.Context, id string, status string, approverID string, notes string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Ambil status lama dan nominal expense
	var currentStatus, accountID string
	var amount int64
	err = tx.QueryRowContext(ctx, "SELECT status, financial_account_id, amount FROM expenses WHERE id = ?;", id).Scan(&currentStatus, &accountID, &amount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrExpenseNotFound
		}
		return err
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// Jika berubah dari non-approved ke approved: potong saldo akun
	if currentStatus != "approved" && status == "approved" {
		_, err = tx.ExecContext(ctx, "UPDATE financial_accounts SET current_balance = current_balance - ?, updated_at = ? WHERE id = ?;", amount, nowStr, accountID)
		if err != nil {
			return fmt.Errorf("potong saldo akun gagal: %w", err)
		}
	}

	// Jika dibatalkan/voided dari sebelumnya approved: kembalikan saldo akun
	if currentStatus == "approved" && (status == "voided" || status == "rejected") {
		_, err = tx.ExecContext(ctx, "UPDATE financial_accounts SET current_balance = current_balance + ?, updated_at = ? WHERE id = ?;", amount, nowStr, accountID)
		if err != nil {
			return fmt.Errorf("restitusi saldo akun gagal: %w", err)
		}
	}

	updateSQL := `
		UPDATE expenses
		SET status = ?, approved_by_user_id = ?, notes = ?, updated_at = ?
		WHERE id = ?;
	`
	_, err = tx.ExecContext(ctx, updateSQL, status, approverID, notes, nowStr, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ==========================================
// 3. LAPORAN LABA RUGI (PROFIT & LOSS)
// ==========================================

func (r *FinanceRepository) GetProfitLossData(ctx context.Context, startDate, endDate string) (*models.ProfitLossReport, error) {
	report := &models.ProfitLossReport{
		PeriodStart: startDate,
		PeriodEnd:   endDate,
	}

	// 1. Agregasi Penjualan (Hanya Order yang Telah 'paid')
	salesQuery := `
		SELECT 
			COALESCE(SUM(subtotal), 0) AS gross_sales,
			COALESCE(SUM(discount_amount), 0) AS discounts,
			COALESCE(SUM(tax_amount), 0) AS tax_collected,
			COALESCE(SUM(total_amount), 0) AS net_sales
		FROM orders
		WHERE payment_status = 'paid'
		  AND date(created_at) >= date(?)
		  AND date(created_at) <= date(?);
	`
	err := r.db.QueryRowContext(ctx, salesQuery, startDate, endDate).Scan(
		&report.GrossSales, &report.Discounts, &report.TaxCollected, &report.NetSales,
	)
	if err != nil {
		return nil, fmt.Errorf("query sales gagal: %w", err)
	}

	// 2. Agregasi HPP (COGS - Cost of Goods Sold) berdasarkan resep menu yang terjual
	cogsQuery := `
		SELECT COALESCE(SUM(oi.quantity * ri.quantity * ing.cost_per_unit), 0) AS total_cogs
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		JOIN recipes rec ON oi.menu_item_id = rec.menu_item_id
		JOIN recipe_items ri ON rec.id = ri.recipe_id
		JOIN ingredients ing ON ri.ingredient_id = ing.id
		WHERE o.payment_status = 'paid'
		  AND date(o.created_at) >= date(?)
		  AND date(o.created_at) <= date(?);
	`
	_ = r.db.QueryRowContext(ctx, cogsQuery, startDate, endDate).Scan(&report.COGS)

	// Laba Kotor = Penjualan Bersih - HPP
	report.GrossProfit = report.NetSales - report.COGS
	if report.NetSales > 0 {
		report.GrossMarginPct = (float64(report.GrossProfit) / float64(report.NetSales)) * 100.0
	}

	// 3. Agregasi Biaya Operasional (Expenses dengan status 'approved')
	expQuery := `
		SELECT ec.name, COALESCE(SUM(e.amount), 0) as cat_total
		FROM expenses e
		JOIN expense_categories ec ON e.category_id = ec.id
		WHERE e.status = 'approved'
		  AND date(e.expense_date) >= date(?)
		  AND date(e.expense_date) <= date(?)
		GROUP BY ec.name
		ORDER BY cat_total DESC;
	`
	rows, err := r.db.QueryContext(ctx, expQuery, startDate, endDate)
	if err == nil {
		defer rows.Close()
		var totalExpenses int64
		for rows.Next() {
			var catName string
			var catAmount int64
			if err := rows.Scan(&catName, &catAmount); err == nil {
				totalExpenses += catAmount
				report.ExpenseBreakdown = append(report.ExpenseBreakdown, &models.ExpenseCategorySummary{
					CategoryName: catName,
					TotalAmount:  catAmount,
				})
			}
		}
		report.OperatingExpenses = totalExpenses

		// Hitung persentase breakdown biaya
		for _, cat := range report.ExpenseBreakdown {
			if totalExpenses > 0 {
				cat.Percentage = (float64(cat.TotalAmount) / float64(totalExpenses)) * 100.0
			}
		}
	}

	// Laba Bersih = Laba Kotor - Biaya Operasional
	report.NetProfit = report.GrossProfit - report.OperatingExpenses
	if report.NetSales > 0 {
		report.NetMarginPct = (float64(report.NetProfit) / float64(report.NetSales)) * 100.0
	}

	return report, nil
}
