package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
	"kopirise/backend/internal/service"
)

// TableHandler menangani endpoint meja, area, dan token QR.
type TableHandler struct {
	tableService *service.TableService
}

// NewTableHandler membuat instance TableHandler.
func NewTableHandler(tableService *service.TableService) *TableHandler {
	return &TableHandler{tableService: tableService}
}

func (h *TableHandler) ListAreas(w http.ResponseWriter, r *http.Request) {
	areas, err := h.tableService.ListAreas(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat area", nil)
		return
	}
	RespondSuccess(w, areas)
}

func (h *TableHandler) ListTables(w http.ResponseWriter, r *http.Request) {
	areaID := r.URL.Query().Get("area_id")
	status := r.URL.Query().Get("status")

	tables, err := h.tableService.ListTables(r.Context(), areaID, status)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat data meja", nil)
		return
	}
	RespondSuccess(w, tables)
}

func (h *TableHandler) GetTable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	table, err := h.tableService.GetTableByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrTableNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Meja tidak ditemukan", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	RespondSuccess(w, table)
}

func (h *TableHandler) CreateTable(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	var req models.CreateTableRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	table, err := h.tableService.CreateTable(r.Context(), &req, actor, getIP(r))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
		return
	}
	RespondCreated(w, table)
}

func (h *TableHandler) UpdateTable(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	var req models.UpdateTableRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	table, err := h.tableService.UpdateTable(r.Context(), id, &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrTableNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Meja tidak ditemukan", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, table)
}

func (h *TableHandler) GenerateQR(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	tokenStr, err := h.tableService.GenerateQRToken(r.Context(), id, actor, getIP(r))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "TOKEN_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, map[string]string{
		"table_id": id,
		"token":    tokenStr,
	})
}

func (h *TableHandler) RevokeQR(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	if err := h.tableService.RevokeQRToken(r.Context(), id, actor, getIP(r)); err != nil {
		RespondError(w, http.StatusInternalServerError, "REVOKE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, map[string]string{
		"message": "Token QR berhasil dicabut.",
	})
}

// ResolvePublicQR menangani scan QR publik dari browser ponsel pelanggan.
func (h *TableHandler) ResolvePublicQR(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	res, err := h.tableService.ResolvePublicQR(r.Context(), token)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidQRToken) {
			RespondError(w, http.StatusNotFound, "INVALID_TOKEN", "QR Code meja tidak valid atau telah dicabut pengelola.", nil)
			return
		}
		if errors.Is(err, repository.ErrTableInactive) {
			RespondError(w, http.StatusForbidden, "TABLE_INACTIVE", "Meja ini sedang tidak aktif menerima pesanan.", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "RESOLVE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, res)
}
