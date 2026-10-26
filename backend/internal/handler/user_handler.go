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

// UserHandler menangani endpoint pengelolaan pengguna dan karyawan.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler membuat instance UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List menampilkan daftar pengguna.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	users, meta, err := h.userService.ListUsers(r.Context(), search, role, status, page, perPage)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar pengguna.", nil)
		return
	}

	RespondPagination(w, users, meta)
}

// Create membuat akun pengguna baru.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreateUserRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	created, err := h.userService.CreateUser(r.Context(), &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			RespondError(w, http.StatusConflict, "EMAIL_EXISTS", "Email ini sudah terdaftar dalam sistem.", nil)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedAction) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Anda tidak memiliki izin membuat akun.", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, created)
}

// Update memperbarui data profil pengguna.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	userID := chi.URLParam(r, "id")

	var req models.UpdateUserRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	updated, err := h.userService.UpdateUser(r.Context(), userID, &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Pengguna tidak ditemukan.", nil)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedAction) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Tidak memiliki izin mengubah akun ini.", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}

	RespondSuccess(w, updated)
}

// ResetPassword mereset password akun pengguna.
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	userID := chi.URLParam(r, "id")

	var req models.ResetPasswordRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.userService.ResetPassword(r.Context(), userID, req.NewPassword, actor, getIP(r)); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Pengguna tidak ditemukan.", nil)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedAction) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Tidak memiliki izin mereset password akun ini.", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "RESET_FAILED", err.Error(), nil)
		return
	}

	RespondSuccess(w, map[string]string{
		"message": "Kata sandi berhasil direset.",
	})
}
