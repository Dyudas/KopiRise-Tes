package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"kopirise/backend/internal/config"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/pkg/upload"
	"kopirise/backend/internal/repository"
	"kopirise/backend/internal/service"
)

// MenuHandler menangani endpoint API untuk kategori dan item menu.
type MenuHandler struct {
	cfg         *config.Config
	menuService *service.MenuService
}

// NewMenuHandler membuat instance MenuHandler baru.
func NewMenuHandler(cfg *config.Config, menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{
		cfg:         cfg,
		menuService: menuService,
	}
}

// Categories
func (h *MenuHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("only_active") == "true"
	categories, err := h.menuService.ListCategories(r.Context(), onlyActive)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat kategori menu", nil)
		return
	}
	RespondSuccess(w, categories)
}

func (h *MenuHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	var req models.CreateCategoryRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	cat, err := h.menuService.CreateCategory(r.Context(), &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrSlugExists) {
			RespondError(w, http.StatusConflict, "SLUG_EXISTS", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
		return
	}
	RespondCreated(w, cat)
}

func (h *MenuHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	var req models.UpdateCategoryRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	cat, err := h.menuService.UpdateCategory(r.Context(), id, &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, cat)
}

func (h *MenuHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	if err := h.menuService.DeleteCategory(r.Context(), id, actor, getIP(r)); err != nil {
		if errors.Is(err, repository.ErrCategoryHasItems) {
			RespondError(w, http.StatusConflict, "CATEGORY_HAS_ITEMS", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, map[string]string{"message": "Kategori berhasil dihapus"})
}

// Menu Items
func (h *MenuHandler) ListMenuItems(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	categoryID := r.URL.Query().Get("category_id")
	availableOnly := r.URL.Query().Get("available_only") == "true"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	items, meta, err := h.menuService.ListMenuItems(r.Context(), search, categoryID, availableOnly, page, perPage)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat item menu", nil)
		return
	}
	RespondPagination(w, items, meta)
}

func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.menuService.GetMenuItemByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrMenuItemNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	RespondSuccess(w, item)
}

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	var req models.CreateMenuItemRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	item, err := h.menuService.CreateMenuItem(r.Context(), &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrSlugExists) {
			RespondError(w, http.StatusConflict, "SLUG_EXISTS", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
		return
	}
	RespondCreated(w, item)
}

func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	var req models.UpdateMenuItemRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	item, err := h.menuService.UpdateMenuItem(r.Context(), id, &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrMenuItemNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, item)
}

func (h *MenuHandler) ToggleAvailability(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	var body struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := DecodeAndValidate(r, &body); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.menuService.ToggleAvailability(r.Context(), id, body.IsAvailable, actor, getIP(r)); err != nil {
		RespondError(w, http.StatusInternalServerError, "TOGGLE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, map[string]interface{}{"id": id, "is_available": body.IsAvailable})
}

func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	if err := h.menuService.DeleteMenuItem(r.Context(), id, actor, getIP(r)); err != nil {
		RespondError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, map[string]string{"message": "Item menu berhasil dinonaktifkan (soft delete)"})
}

// UploadPhoto memproses unggahan gambar foto menu dengan validasi magic bytes.
func (h *MenuHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(10 << 20) // 10 MB

	file, header, err := r.FormFile("photo")
	if err != nil {
		RespondError(w, http.StatusBadRequest, "FILE_REQUIRED", "File foto harus dilampirkan", nil)
		return
	}
	defer file.Close()

	uploadDir := "data/uploads"
	res, err := upload.SaveImage(uploadDir, header.Filename, file)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "UPLOAD_FAILED", err.Error(), nil)
		return
	}

	url := "/uploads/" + res.StoredName
	RespondCreated(w, map[string]interface{}{
		"url":         url,
		"file_name":   res.StoredName,
		"file_size":   res.FileSize,
		"mime_type":   res.MimeType,
	})
}
