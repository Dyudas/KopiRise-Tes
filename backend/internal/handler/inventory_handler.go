package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
	"kopirise/backend/internal/service"
)

type InventoryHandler struct {
	inventoryService *service.InventoryService
}

func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// 1. Bahan Baku (Ingredients)
func (h *InventoryHandler) ListIngredients(w http.ResponseWriter, r *http.Request) {
	list, err := h.inventoryService.ListIngredients(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar bahan baku", nil)
		return
	}
	RespondSuccess(w, list)
}

func (h *InventoryHandler) ListLowStock(w http.ResponseWriter, r *http.Request) {
	list, err := h.inventoryService.ListLowStock(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat peringatan stok menipis", nil)
		return
	}
	RespondSuccess(w, list)
}

func (h *InventoryHandler) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreateIngredientRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	ing, err := h.inventoryService.CreateIngredient(r.Context(), &req, actor, getIP(r))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "CREATION_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, ing)
}

func (h *InventoryHandler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	var req models.CreateIngredientRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.inventoryService.UpdateIngredient(r.Context(), id, &req, actor, getIP(r)); err != nil {
		if errors.Is(err, repository.ErrIngredientNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Bahan baku tidak ditemukan", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}

	RespondSuccess(w, map[string]string{"id": id, "status": "updated"})
}

// 2. Suppliers
func (h *InventoryHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	list, err := h.inventoryService.ListSuppliers(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat supplier", nil)
		return
	}
	RespondSuccess(w, list)
}

func (h *InventoryHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreateSupplierRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	sup, err := h.inventoryService.CreateSupplier(r.Context(), &req, actor, getIP(r))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "CREATION_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, sup)
}

// 3. Pengadaan & Pembelian Bahan (Purchases)
func (h *InventoryHandler) CreatePurchase(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreatePurchaseRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	po, err := h.inventoryService.CreatePurchase(r.Context(), &req, actor, getIP(r))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "PURCHASE_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, po)
}

func (h *InventoryHandler) ListPurchases(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	purchases, meta, err := h.inventoryService.ListPurchases(r.Context(), page, perPage)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat riwayat pembelian", nil)
		return
	}
	RespondPagination(w, purchases, meta)
}

// 4. Resep Menu (Bill of Materials) & HPP
func (h *InventoryHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	menuItemID := chi.URLParam(r, "menuItemId")
	rec, err := h.inventoryService.GetRecipe(r.Context(), menuItemID)
	if err != nil {
		if errors.Is(err, repository.ErrRecipeNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Resep menu belum dikonfigurasi", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	RespondSuccess(w, rec)
}

func (h *InventoryHandler) SaveRecipe(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreateRecipeRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	rec, err := h.inventoryService.SaveRecipe(r.Context(), &req, actor, getIP(r))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "SAVE_RECIPE_FAILED", err.Error(), nil)
		return
	}

	RespondSuccess(w, rec)
}
