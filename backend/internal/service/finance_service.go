package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
)

var (
	ErrUnauthorizedExpenseAction = errors.New("hanya Admin dan Manager yang berhak menyetujui biaya operasional")
)

type FinanceService struct {
	financeRepo *repository.FinanceRepository
	auditRepo   *repository.AuditRepository
}

func NewFinanceService(financeRepo *repository.FinanceRepository, auditRepo *repository.AuditRepository) *FinanceService {
	return &FinanceService{
		financeRepo: financeRepo,
		auditRepo:   auditRepo,
	}
}

// 1. Bagan Akun Keuangan
func (s *FinanceService) ListAccounts(ctx context.Context) ([]*models.FinancialAccount, error) {
	return s.financeRepo.ListAccounts(ctx)
}

func (s *FinanceService) GetAccountByID(ctx context.Context, id string) (*models.FinancialAccount, error) {
	return s.financeRepo.GetAccountByID(ctx, id)
}

func (s *FinanceService) CreateTransfer(ctx context.Context, req *models.CreateTransferRequest, actor *models.User, ip string) (*models.AccountTransfer, error) {
	transfer := &models.AccountTransfer{
		ID:              "trf-" + generateRandomHex(4),
		TransferNumber:  generateTransferNumber(),
		FromAccountID:   req.FromAccountID,
		ToAccountID:     req.ToAccountID,
		Amount:          req.Amount,
		TransferDate:    req.TransferDate,
		Description:     req.Description,
		CreatedByUserID: actor.ID,
		CreatedAt:       time.Now().UTC(),
	}

	if err := s.financeRepo.CreateTransfer(ctx, transfer); err != nil {
		return nil, err
	}

	// Audit Log
	_ = s.auditRepo.Log(ctx, &models.ActivityLog{
		UserID:       &actor.ID,
		Action:       "CREATE_TRANSFER",
		ResourceType: "account_transfers",
		ResourceID:   &transfer.ID,
		Details:      fmt.Sprintf("Transfer dana Rp %d (#%s) dari akun %s ke %s", transfer.Amount, transfer.TransferNumber, transfer.FromAccountID, transfer.ToAccountID),
		IPAddress:    ip,
	})

	return transfer, nil
}

func (s *FinanceService) ListTransfers(ctx context.Context, page, perPage int) ([]*models.AccountTransfer, *models.PaginationMeta, error) {
	return s.financeRepo.ListTransfers(ctx, page, perPage)
}

// 2. Biaya Operasional (Expenses)
func (s *FinanceService) ListCategories(ctx context.Context) ([]*models.ExpenseCategory, error) {
	return s.financeRepo.ListExpenseCategories(ctx)
}

func (s *FinanceService) CreateExpense(ctx context.Context, req *models.CreateExpenseRequest, attachments []*models.ExpenseAttachment, actor *models.User, ip string) (*models.Expense, error) {
	// Status default: jika dibuat oleh Admin/Manager langsung 'approved', jika staff lain 'pending'
	status := "pending"
	var approvedByID *string
	if actor.HasAnyRole(models.RoleAdmin, models.RoleManager) {
		status = "approved"
		approvedByID = &actor.ID
	}

	exp := &models.Expense{
		ID:                 "exp-" + generateRandomHex(4),
		ExpenseNumber:      generateExpenseNumber(),
		ExpenseDate:        req.ExpenseDate,
		CategoryID:         req.CategoryID,
		FinancialAccountID: req.FinancialAccountID,
		Amount:             req.Amount,
		Description:        req.Description,
		Vendor:             req.Vendor,
		ReferenceNumber:    req.ReferenceNumber,
		Status:             status,
		CreatedByUserID:    actor.ID,
		ApprovedByUserID:   approvedByID,
		Notes:              req.Notes,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	if err := s.financeRepo.CreateExpense(ctx, exp, attachments); err != nil {
		return nil, err
	}

	_ = s.auditRepo.Log(ctx, &models.ActivityLog{
		UserID:       &actor.ID,
		Action:       "CREATE_EXPENSE",
		ResourceType: "expenses",
		ResourceID:   &exp.ID,
		Details:      fmt.Sprintf("Mencatat pengeluaran #%s sebesar Rp %d: %s", exp.ExpenseNumber, exp.Amount, exp.Description),
		IPAddress:    ip,
	})

	return exp, nil
}

func (s *FinanceService) ListExpenses(ctx context.Context, status, categoryID, startDate, endDate string, page, perPage int) ([]*models.Expense, *models.PaginationMeta, error) {
	return s.financeRepo.ListExpenses(ctx, status, categoryID, startDate, endDate, page, perPage)
}

func (s *FinanceService) GetExpenseByID(ctx context.Context, id string) (*models.Expense, error) {
	return s.financeRepo.GetExpenseByID(ctx, id)
}

func (s *FinanceService) UpdateExpenseStatus(ctx context.Context, id, status, notes string, actor *models.User, ip string) error {
	if !actor.HasAnyRole(models.RoleAdmin, models.RoleManager) {
		return ErrUnauthorizedExpenseAction
	}

	if err := s.financeRepo.UpdateExpenseStatus(ctx, id, status, actor.ID, notes); err != nil {
		return err
	}

	_ = s.auditRepo.Log(ctx, &models.ActivityLog{
		UserID:       &actor.ID,
		Action:       "UPDATE_EXPENSE_STATUS",
		ResourceType: "expenses",
		ResourceID:   &id,
		Details:      fmt.Sprintf("Update status pengeluaran ID %s menjadi %s", id, status),
		IPAddress:    ip,
	})

	return nil
}

// 3. Laporan Laba Rugi
func (s *FinanceService) GetProfitLossReport(ctx context.Context, startDate, endDate string) (*models.ProfitLossReport, error) {
	if startDate == "" {
		startDate = time.Now().Format("2006-01-01") // Awal tahun berjalan
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02") // Hari ini
	}

	return s.financeRepo.GetProfitLossData(ctx, startDate, endDate)
}

// Helper Generators
func generateTransferNumber() string {
	today := time.Now().Format("20060102")
	return fmt.Sprintf("TRF-%s-%s", today, strings.ToUpper(generateRandomHex(2)))
}

func generateExpenseNumber() string {
	today := time.Now().Format("20060102")
	return fmt.Sprintf("EXP-%s-%s", today, strings.ToUpper(generateRandomHex(2)))
}

func generateRandomHex(byteLength int) string {
	b := make([]byte, byteLength)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
