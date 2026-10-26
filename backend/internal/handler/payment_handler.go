package handler

import (
	"errors"
	"net/http"

	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
	"kopirise/backend/internal/service"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// ProcessPayment memproses pembayaran kasir atomik.
func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreatePaymentRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	payment, err := h.paymentService.ProcessPayment(r.Context(), &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicatePaymentIdem) {
			RespondError(w, http.StatusConflict, "DUPLICATE_PAYMENT", err.Error(), nil)
			return
		}
		if errors.Is(err, repository.ErrOverpayment) || errors.Is(err, repository.ErrOrderAlreadyPaid) {
			RespondError(w, http.StatusBadRequest, "INVALID_PAYMENT", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "PAYMENT_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, payment)
}

// ProcessRefund memproses refund ter-audit (hanya admin & manager).
func (h *PaymentHandler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreateRefundRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	refund, err := h.paymentService.ProcessRefund(r.Context(), &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, service.ErrUnauthorizedAction) {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "Hanya Admin dan Manager yang berhak menyetujui refund", nil)
			return
		}
		if errors.Is(err, repository.ErrRefundExceedsPaid) {
			RespondError(w, http.StatusBadRequest, "INVALID_REFUND", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "REFUND_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, refund)
}

// Cash Shifts
func (h *PaymentHandler) OpenShift(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.OpenShiftRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	shift, err := h.paymentService.OpenShift(r.Context(), actor.ID, &req, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrActiveShiftExists) {
			RespondError(w, http.StatusConflict, "ACTIVE_SHIFT_EXISTS", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "OPEN_SHIFT_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, shift)
}

func (h *PaymentHandler) CloseShift(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CloseShiftRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	shift, err := h.paymentService.CloseShift(r.Context(), actor.ID, &req, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveShift) {
			RespondError(w, http.StatusNotFound, "NO_ACTIVE_SHIFT", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "CLOSE_SHIFT_FAILED", err.Error(), nil)
		return
	}

	RespondSuccess(w, shift)
}

func (h *PaymentHandler) CreateMovement(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	var req models.CreateCashMovementRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	mov, err := h.paymentService.CreateCashMovement(r.Context(), actor.ID, &req, getIP(r))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "MOVEMENT_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, mov)
}

func (h *PaymentHandler) GetCurrentShift(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)

	shift, err := h.paymentService.GetCurrentShift(r.Context(), actor.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveShift) {
			RespondError(w, http.StatusNotFound, "NO_ACTIVE_SHIFT", "Kasir tidak memiliki shift kas yang sedang aktif", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	RespondSuccess(w, shift)
}
