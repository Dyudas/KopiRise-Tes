package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
)

type PaymentService struct {
	paymentRepo *repository.PaymentRepository
	orderRepo   *repository.OrderRepository
	audit       *repository.AuditRepository
}

func NewPaymentService(paymentRepo *repository.PaymentRepository, orderRepo *repository.OrderRepository, audit *repository.AuditRepository) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		audit:       audit,
	}
}

// ProcessPayment memproses pembayaran kasir secara atomik dengan kalkulasi kembalian uang tunai.
func (s *PaymentService) ProcessPayment(ctx context.Context, req *models.CreatePaymentRequest, actor *models.User, ip string) (*models.Payment, error) {
	// Ambil order
	order, err := s.orderRepo.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	remaining := order.TotalAmount - order.PaidAmount
	if remaining <= 0 {
		return nil, repository.ErrOrderAlreadyPaid
	}
	if req.Amount > remaining {
		return nil, repository.ErrOverpayment
	}

	var changeDue *int64
	if req.PaymentMethod == "cash" {
		if req.CashTendered == nil || *req.CashTendered < req.Amount {
			return nil, errors.New("uang tunai yang diterima kurang dari nominal tagihan")
		}
		chg := *req.CashTendered - req.Amount
		changeDue = &chg
	}

	// Cek shift kas aktif kasir jika tunai
	var shiftID *string
	if req.PaymentMethod == "cash" {
		activeShift, err := s.paymentRepo.GetActiveCashShift(ctx, actor.ID)
		if err != nil {
			return nil, fmt.Errorf("pembayaran tunai memerlukan shift kasir aktif: %w", err)
		}
		shiftID = &activeShift.ID
	}

	now := time.Now().UTC()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	datePart := now.In(loc).Format("20060102")
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randNum := n.Int64() + 1000
	payNumber := fmt.Sprintf("PAY-%s-%04d", datePart, randNum)

	payID := "pay-" + uuid.NewString()[:8]
	payment := &models.Payment{
		ID:                 payID,
		OrderID:            order.ID,
		CashShiftID:        shiftID,
		PaymentNumber:      payNumber,
		IdempotencyKey:     &req.IdempotencyKey,
		PaymentMethod:      req.PaymentMethod,
		Amount:             req.Amount,
		CashTendered:       req.CashTendered,
		ChangeDue:          changeDue,
		ReferenceNumber:    req.ReferenceNumber,
		ProcessedByUserID:  actor.ID,
		FinancialAccountID: req.FinancialAccountID,
		Status:             "completed",
		Notes:              req.Notes,
		CreatedAt:          now,
	}

	if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "payment.create",
		EntityType: "payment",
		EntityID:   &payment.ID,
		Details:    fmt.Sprintf(`{"number":"%s","amount":%d,"method":"%s"}`, payment.PaymentNumber, payment.Amount, payment.PaymentMethod),
		IPAddress:  ip,
	})

	return payment, nil
}

// ProcessRefund memproses refund ter-audit (hanya admin/manager).
func (s *PaymentService) ProcessRefund(ctx context.Context, req *models.CreateRefundRequest, actor *models.User, ip string) (*models.Refund, error) {
	if !actor.HasAnyRole(models.RoleAdmin, models.RoleManager) {
		return nil, ErrUnauthorizedAction
	}

	now := time.Now().UTC()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	datePart := now.In(loc).Format("20060102")
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randNum := n.Int64() + 1000
	refNumber := fmt.Sprintf("REF-%s-%04d", datePart, randNum)

	// Cek shift kas aktif jika refund tunai
	var shiftID *string
	if req.RefundMethod == "cash" {
		activeShift, err := s.paymentRepo.GetActiveCashShift(ctx, actor.ID)
		if err == nil {
			shiftID = &activeShift.ID
		}
	}

	refund := &models.Refund{
		ID:                 "ref-" + uuid.NewString()[:8],
		RefundNumber:       refNumber,
		PaymentID:          req.PaymentID,
		OrderID:            req.OrderID,
		CashShiftID:        shiftID,
		Amount:             req.Amount,
		Reason:             req.Reason,
		RefundMethod:       req.RefundMethod,
		FinancialAccountID: req.FinancialAccountID,
		ApprovedByUserID:   actor.ID,
		CreatedAt:          now,
	}

	if err := s.paymentRepo.CreateRefund(ctx, refund); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "payment.refund",
		EntityType: "refund",
		EntityID:   &refund.ID,
		Details:    fmt.Sprintf(`{"refund_number":"%s","amount":%d,"reason":"%s"}`, refund.RefundNumber, refund.Amount, refund.Reason),
		IPAddress:  ip,
	})

	return refund, nil
}

// Cash Shift

func (s *PaymentService) OpenShift(ctx context.Context, cashierID string, req *models.OpenShiftRequest, ip string) (*models.CashShift, error) {
	now := time.Now().UTC()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	datePart := now.In(loc).Format("20060102")
	n, _ := rand.Int(rand.Reader, big.NewInt(90))
	shiftNumber := fmt.Sprintf("SHF-%s-%02d", datePart, n.Int64()+10)

	shf := &models.CashShift{
		ID:            "shf-" + uuid.NewString()[:8],
		CashierUserID: cashierID,
		ShiftNumber:   shiftNumber,
		InitialCash:   req.InitialCash,
		OpenedAt:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.paymentRepo.OpenCashShift(ctx, shf); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &cashierID,
		Action:     "cash_shift.open",
		EntityType: "cash_shift",
		EntityID:   &shf.ID,
		Details:    fmt.Sprintf(`{"shift_number":"%s","initial_cash":%d}`, shf.ShiftNumber, shf.InitialCash),
		IPAddress:  ip,
	})

	return shf, nil
}

func (s *PaymentService) CloseShift(ctx context.Context, cashierID string, req *models.CloseShiftRequest, ip string) (*models.CashShift, error) {
	current, err := s.paymentRepo.GetActiveCashShift(ctx, cashierID)
	if err != nil {
		return nil, err
	}

	// Formula: kas_seharusnya = kas_awal + penjualan_tunai + kas_masuk - kas_keluar - refund_tunai
	expected := current.InitialCash + current.CashSales + current.CashIn - current.CashOut - current.CashRefunds
	discrepancy := req.ActualCash - expected

	if err := s.paymentRepo.CloseCashShift(ctx, current.ID, req.ActualCash, discrepancy, req.ClosingNotes); err != nil {
		return nil, err
	}

	current.Status = "closed"
	current.ActualCash = &req.ActualCash
	current.Discrepancy = &discrepancy
	current.ClosingNotes = req.ClosingNotes

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &cashierID,
		Action:     "cash_shift.close",
		EntityType: "cash_shift",
		EntityID:   &current.ID,
		Details:    fmt.Sprintf(`{"expected":%d,"actual":%d,"discrepancy":%d}`, expected, req.ActualCash, discrepancy),
		IPAddress:  ip,
	})

	return current, nil
}

func (s *PaymentService) CreateCashMovement(ctx context.Context, cashierID string, req *models.CreateCashMovementRequest, ip string) (*models.CashMovement, error) {
	current, err := s.paymentRepo.GetActiveCashShift(ctx, cashierID)
	if err != nil {
		return nil, err
	}

	mov := &models.CashMovement{
		ID:                 "mov-" + uuid.NewString()[:8],
		CashShiftID:        current.ID,
		Type:               req.Type,
		Amount:             req.Amount,
		Reason:             req.Reason,
		AuthorizedByUserID: &cashierID,
		CreatedAt:          time.Now().UTC(),
	}

	if err := s.paymentRepo.CreateCashMovement(ctx, mov); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &cashierID,
		Action:     "cash_shift.movement",
		EntityType: "cash_movement",
		EntityID:   &mov.ID,
		Details:    fmt.Sprintf(`{"type":"%s","amount":%d,"reason":"%s"}`, mov.Type, mov.Amount, mov.Reason),
		IPAddress:  ip,
	})

	return mov, nil
}

func (s *PaymentService) GetCurrentShift(ctx context.Context, cashierID string) (*models.CashShift, error) {
	return s.paymentRepo.GetActiveCashShift(ctx, cashierID)
}
