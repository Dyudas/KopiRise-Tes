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

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrder menangani pembuatan pesanan dari POS kasir maupun QR order meja publik.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	var actor *models.User
	if u, ok := r.Context().Value(models.ContextKeyUser).(*models.User); ok {
		actor = u
	}

	order, err := h.orderService.CreateOrder(r.Context(), &req, actor, getIP(r))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateIdempotency) {
			RespondError(w, http.StatusConflict, "DUPLICATE_REQUEST", err.Error(), nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "ORDER_CREATION_FAILED", err.Error(), nil)
		return
	}

	RespondCreated(w, order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	order, err := h.orderService.GetOrderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Pesanan tidak ditemukan", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	RespondSuccess(w, order)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	paymentStatus := r.URL.Query().Get("payment_status")
	orderType := r.URL.Query().Get("order_type")
	dateStr := r.URL.Query().Get("date")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	orders, meta, err := h.orderService.ListOrders(r.Context(), status, paymentStatus, orderType, dateStr, page, perPage)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat pesanan", nil)
		return
	}
	RespondPagination(w, orders, meta)
}

// ListKitchenOrders antrean pesanan khusus Kitchen Display System (Barista & Dapur).
func (h *OrderHandler) ListKitchenOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orderService.ListKitchenOrders(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat antrean dapur", nil)
		return
	}
	RespondSuccess(w, orders)
}

// UpdateOrderStatus memperbarui status pesanan (misal: preparing, ready, served, completed).
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	actor, _ := r.Context().Value(models.ContextKeyUser).(*models.User)
	id := chi.URLParam(r, "id")

	var body struct {
		Status string `json:"status" validate:"required,oneof=pending confirmed preparing ready served completed cancelled"`
	}
	if err := DecodeAndValidate(r, &body); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.orderService.UpdateOrderStatus(r.Context(), id, body.Status, actor, getIP(r)); err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "Pesanan tidak ditemukan", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, map[string]string{"id": id, "status": body.Status})
}

// UpdateOrderItemStatus memperbarui status per item masakan/minuman di KDS.
func (h *OrderHandler) UpdateOrderItemStatus(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")

	var body struct {
		Status string `json:"status" validate:"required,oneof=pending preparing ready served cancelled"`
	}
	if err := DecodeAndValidate(r, &body); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.orderService.UpdateOrderItemStatus(r.Context(), itemID, body.Status); err != nil {
		RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}
	RespondSuccess(w, map[string]string{"item_id": itemID, "status": body.Status})
}

// GetReceipt Data struk terformat untuk printer thermal 58mm/80mm.
func (h *OrderHandler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	order, err := h.orderService.GetOrderByID(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Pesanan tidak ditemukan", nil)
		return
	}

	receiptData := map[string]interface{}{
		"cafe_name":    "KopiRise",
		"cafe_address": "Jl. Semeru, Kloncing, Sumbersari, Jember",
		"cafe_phone":   "081234567890",
		"order_number": order.OrderNumber,
		"order_type":   order.OrderType,
		"table_number": order.TableNumber,
		"date":         order.CreatedAt.Format("02/01/2006 15:04"),
		"items":        order.Items,
		"subtotal":     order.Subtotal,
		"discount":     order.DiscountAmount,
		"tax":          order.TaxAmount,
		"total":        order.TotalAmount,
		"paid_amount":  order.PaidAmount,
		"status":       order.PaymentStatus,
		"wifi_ssid":    "KopiRise_Guest",
		"wifi_pass":    "semerujember",
	}

	RespondSuccess(w, receiptData)
}
