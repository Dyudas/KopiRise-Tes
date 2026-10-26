package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/pkg/upload"
	"kopirise/backend/internal/repository"
	"kopirise/backend/internal/service"
)

type FinanceHandler struct {
	financeService   *service.FinanceService
	inventoryService *service.InventoryService
	exportService    *service.ExportService
}

func NewFinanceHandler(
	financeService *service.FinanceService,
	inventoryService *service.InventoryService,
	exportService *service.ExportService,
) *FinanceHandler {
	return &FinanceHandler{
		financeService:   financeService,
		inventoryService: inventoryService,
		exportService:    exportService,
	}
}

// 1. Akun Keuangan & Transfer
func (h *FinanceHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.financeService.ListAccounts(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat akun keuangan", nil)
		return
	}
	RespondSuccess(w, accounts)
}

func (h *FinanceHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreateTransferRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	trf, err := h.financeService.CreateTransfer(r.Context(), &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) || errors.Is(err, repository.ErrSameAccountTransfer) {
			RespondError(w, http.StatusBadRequest, "INVALID_TRANSFER", err.Error(), nil)
			return
		}
		if errors.Is(err, repository.ErrAccountNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "TRANSFER_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, trf)
}

func (h *FinanceHandler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	transfers, meta, err := h.financeService.ListTransfers(r.Context(), page, perPage)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat riwayat transfer", nil)
		return
	}
	RespondPagination(w, transfers, meta)
}

// 2. Kategori & Biaya Operasional (Expenses)
func (h *FinanceHandler) ListExpenseCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.financeService.ListCategories(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat kategori biaya", nil)
		return
	}
	RespondSuccess(w, cats)
}

func (h *FinanceHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	// Jika multipart/form-data: tangani upload file bukti/nota
	var attachments []*models.ExpenseAttachment
	var req models.CreateExpenseRequest

	if r.Header.Get("Content-Type") != "" && len(r.Header.Get("Content-Type")) >= 19 && r.Header.Get("Content-Type")[:19] == "multipart/form-data" {
		if err := r.ParseMultipartForm(5 << 20); err != nil { // 5MB limit
			RespondError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "Ukuran form/file melebihi batas 5MB", nil)
			return
		}

		amountInt, _ := strconv.ParseInt(r.FormValue("amount"), 10, 64)
		req = models.CreateExpenseRequest{
			ExpenseDate:        r.FormValue("expense_date"),
			CategoryID:         r.FormValue("category_id"),
			FinancialAccountID: r.FormValue("financial_account_id"),
			Amount:             amountInt,
			Description:        r.FormValue("description"),
			Vendor:             r.FormValue("vendor"),
			ReferenceNumber:    r.FormValue("reference_number"),
			Notes:              r.FormValue("notes"),
		}

		// Handle file upload jika ada
		file, header, err := r.FormFile("attachment")
		if err == nil {
			defer file.Close()
			savedPath, mimeType, err := upload.SaveUploadedImage(file, header, "data/uploads/receipts")
			if err == nil {
				attachments = append(attachments, &models.ExpenseAttachment{
					ID:        "att-" + strconv.FormatInt(time.Now().UnixNano(), 10),
					FilePath:  savedPath,
					FileName:  header.Filename,
					FileSize:  header.Size,
					MimeType:  mimeType,
					CreatedAt: time.Now().UTC(),
				})
			}
		}
	} else {
		if err := DecodeAndValidate(r, &req); err != nil {
			RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
	}

	exp, err := h.financeService.CreateExpense(r.Context(), &req, attachments, actor, getIP(r))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "EXPENSE_CREATION_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, exp)
}

func (h *FinanceHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	categoryID := r.URL.Query().Get("category_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	expenses, meta, err := h.financeService.ListExpenses(r.Context(), status, categoryID, startDate, endDate, page, perPage)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar biaya", nil)
		return
	}
	RespondPagination(w, expenses, meta)
}

func (h *FinanceHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	exp, err := h.financeService.GetExpenseByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrExpenseNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Pengeluaran tidak ditemukan", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	RespondSuccess(w, exp)
}

func (h *FinanceHandler) UpdateExpenseStatus(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	var req models.ApproveExpenseRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.financeService.UpdateExpenseStatus(r.Context(), id, req.Status, req.Notes, actor, getIP(r)); err != nil {
		if errors.Is(err, service.ErrUnauthorizedExpenseAction) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}

	RespondSuccess(w, map[string]string{"id": id, "status": req.Status})
}

// 3. Laporan Laba Rugi & Ekspor
func (h *FinanceHandler) GetProfitLossReport(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	report, err := h.financeService.GetProfitLossReport(r.Context(), startDate, endDate)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "REPORT_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, report)
}

func (h *FinanceHandler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	report, err := h.financeService.GetProfitLossReport(r.Context(), startDate, endDate)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "REPORT_FAILED", err.Error(), nil)
		return
	}

	accounts, _ := h.financeService.ListAccounts(r.Context())
	ingredients, _ := h.inventoryService.ListIngredients(r.Context())

	fileBytes, err := h.exportService.ExportProfitLossExcel(report, accounts, ingredients)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "EXPORT_FAILED", err.Error(), nil)
		return
	}

	fileName := fmt.Sprintf("Laporan_Keuangan_KopiRise_%s_%s.xlsx", report.PeriodStart, report.PeriodEnd)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fileBytes)
}

func (h *FinanceHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	report, err := h.financeService.GetProfitLossReport(r.Context(), startDate, endDate)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "REPORT_FAILED", err.Error(), nil)
		return
	}

	fileBytes, err := h.exportService.ExportProfitLossCSV(report)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "EXPORT_FAILED", err.Error(), nil)
		return
	}

	fileName := fmt.Sprintf("Laba_Rugi_KopiRise_%s.csv", report.PeriodEnd)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fileBytes)
}
