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

// ReservationHandler menangani endpoint API reservasi meja.
type ReservationHandler struct {
	reservationService *service.ReservationService
}

// NewReservationHandler membuat instance ReservationHandler baru.
func NewReservationHandler(reservationService *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{reservationService: reservationService}
}

// CreatePublicReservation menangani form reservasi dari website publik.
func (h *ReservationHandler) CreatePublicReservation(w http.ResponseWriter, r *http.Request) {
	var req models.CreateReservationRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	res, err := h.reservationService.CreatePublicReservation(r.Context(), &req, getIP(r))
	if err != nil {
		if errors.Is(err, service.ErrSpamDetected) {
			RespondError(w, http.StatusBadRequest, "SPAM_REJECTED", "Permintaan reservasi ditolak.", nil)
			return
		}
		if errors.Is(err, service.ErrPastDate) || errors.Is(err, service.ErrOutsideHours) {
			RespondError(w, http.StatusBadRequest, "INVALID_SCHEDULE", err.Error(), nil)
			return
		}
		if errors.Is(err, repository.ErrDoubleBooking) {
			RespondError(w, http.StatusConflict, "DOUBLE_BOOKING", "Meja yang dipilih sudah terisi pada rentang waktu tersebut. Silakan pilih meja atau jam lain.", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, res)
}

// GetByNumber mengecek status reservasi pelanggan menggunakan nomor unik (misal: RSV-20261026-0001).
func (h *ReservationHandler) GetByNumber(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")
	res, err := h.reservationService.GetReservationByNumber(r.Context(), number)
	if err != nil {
		if errors.Is(err, repository.ErrReservationNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Nomor reservasi tidak ditemukan dalam sistem kami.", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	RespondSuccess(w, res)
}

// ListReservations menampilkan daftar reservasi untuk dashboard manager/admin.
func (h *ReservationHandler) ListReservations(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	list, meta, err := h.reservationService.ListReservations(r.Context(), dateStr, status, page, perPage)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat reservasi", nil)
		return
	}
	RespondPagination(w, list, meta)
}

// UpdateStatus mengubah status reservasi (pending -> confirmed -> seated -> completed / cancelled / no_show).
func (h *ReservationHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	var req models.UpdateReservationStatusRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	updated, err := h.reservationService.UpdateStatus(r.Context(), id, req.Status, req.Notes, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrReservationNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Reservasi tidak ditemukan", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, updated)
}
