package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
)

type OrderService struct {
	orderRepo *repository.OrderRepository
	menuRepo  *repository.MenuRepository
	tableRepo *repository.TableRepository
	audit     *repository.AuditRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, menuRepo *repository.MenuRepository, tableRepo *repository.TableRepository, audit *repository.AuditRepository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		menuRepo:  menuRepo,
		tableRepo: tableRepo,
		audit:     audit,
	}
}

// CreateOrder memproses order baru dengan kalkulasi harga di server dan snapshotting lengkap.
func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest, actor *models.User, ip string) (*models.Order, error) {
	// 1. Idempotency Check: cegah pemrosesan ganda
	if req.IdempotencyKey != "" {
		existing, err := s.orderRepo.GetOrderByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil // Return order yang sudah tersimpan
		}
	}

	// 2. Validasi Meja jika Dine-In
	if req.OrderType == "dine_in" && (req.TableID == nil || *req.TableID == "") {
		return nil, errors.New("meja wajib dipilih untuk pesanan dine-in")
	}

	orderID := "ord-" + uuid.NewString()[:8]
	var orderItems []*models.OrderItem
	var calculatedSubtotal int64 = 0

	now := time.Now().UTC()

	// 3. Proses Item & Snapshots
	for idx, itmReq := range req.Items {
		menuItem, err := s.menuRepo.GetMenuItemByID(ctx, itmReq.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("item menu %s tidak valid: %w", itmReq.MenuItemID, err)
		}
		if !menuItem.IsAvailable {
			return nil, fmt.Errorf("menu %s (%s) sedang habis/tidak tersedia", menuItem.Name, menuItem.ID)
		}

		unitPrice := menuItem.BasePrice
		variantName := ""

		if itmReq.MenuVariantID != nil && *itmReq.MenuVariantID != "" {
			variants, err := s.menuRepo.GetVariantsByMenuItemID(ctx, menuItem.ID)
			if err == nil {
				for _, v := range variants {
					if v.ID == *itmReq.MenuVariantID && v.IsAvailable {
						unitPrice += v.PriceAdjustment
						variantName = v.Name
						break
					}
				}
			}
		}

		itemSubtotal := unitPrice * int64(itmReq.Quantity)
		orderItemID := fmt.Sprintf("itm-%s-%d", orderID[4:], idx+1)

		var itemModifiers []*models.OrderItemModifier
		// Ambil opsi modifier yang dipilih
		if len(itmReq.ModifierOptionIDs) > 0 {
			groups, _ := s.menuRepo.GetModifierGroupsByMenuItemID(ctx, menuItem.ID)
			for _, optID := range itmReq.ModifierOptionIDs {
				for _, g := range groups {
					for _, opt := range g.Options {
						if opt.ID == optID && opt.IsAvailable {
							modID := fmt.Sprintf("mod-%s-%s", orderItemID, opt.ID)
							itemModifiers = append(itemModifiers, &models.OrderItemModifier{
								ID:                      modID,
								OrderItemID:             orderItemID,
								ModifierOptionID:        &opt.ID,
								GroupNameSnapshot:       g.Name,
								OptionNameSnapshot:      opt.Name,
								AdditionalPriceSnapshot: opt.AdditionalPrice,
							})
							itemSubtotal += opt.AdditionalPrice * int64(itmReq.Quantity)
						}
					}
				}
			}
		}

		calculatedSubtotal += itemSubtotal

		orderItems = append(orderItems, &models.OrderItem{
			ID:                    orderItemID,
			OrderID:               orderID,
			MenuItemID:            &menuItem.ID,
			MenuVariantID:         itmReq.MenuVariantID,
			ItemNameSnapshot:      menuItem.Name,
			VariantNameSnapshot:   variantName,
			UnitBasePriceSnapshot: unitPrice,
			Quantity:              itmReq.Quantity,
			Subtotal:              itemSubtotal,
			Notes:                 itmReq.Notes,
			Status:                "pending",
			Modifiers:             itemModifiers,
			CreatedAt:             now,
		})
	}

	// 4. Kalkulasi Diskon Promo (Diskon hanya dihitung di backend!)
	var discountAmount int64 = 0
	var appliedPromoID *string
	if strings.TrimSpace(req.PromoCode) != "" {
		code := strings.ToUpper(strings.TrimSpace(req.PromoCode))
		if code == "SENINSEMANGAT" && calculatedSubtotal >= 40000 {
			discountAmount = int64(math.Floor(float64(calculatedSubtotal) * 0.15))
			if discountAmount > 25000 {
				discountAmount = 25000
			}
			pID := "prm-senin"
			appliedPromoID = &pID
		} else if code == "KOPIPAHAM" && calculatedSubtotal >= 60000 {
			discountAmount = 10000
			pID := "prm-ngopi"
			appliedPromoID = &pID
		}
	}

	// 5. Kalkulasi Pajak PB1 (10%) pada DPP (Dasar Pengenaan Pajak)
	dpp := calculatedSubtotal - discountAmount
	if dpp < 0 {
		dpp = 0
	}
	taxAmount := int64(math.Round(float64(dpp) * 0.10)) // Pajak Restoran PB1 10%
	var serviceCharge int64 = 0
	totalAmount := dpp + serviceCharge + taxAmount

	// 6. Generate Nomor Pesanan: ORD-YYYYMMDD-XXXX
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	nowLocal := now.In(loc)
	datePart := nowLocal.Format("20060102")
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randNum := n.Int64() + 1000
	orderNumber := fmt.Sprintf("ORD-%s-%04d", datePart, randNum)

	var creatorID *string
	if actor != nil {
		creatorID = &actor.ID
	}

	order := &models.Order{
		ID:                  orderID,
		OrderNumber:         orderNumber,
		IdempotencyKey:      &req.IdempotencyKey,
		OrderType:           req.OrderType,
		TableID:             req.TableID,
		CustomerName:        strings.TrimSpace(req.CustomerName),
		CustomerPhone:       strings.TrimSpace(req.CustomerPhone),
		Status:              "confirmed",
		PaymentStatus:       "unpaid",
		Subtotal:            calculatedSubtotal,
		DiscountAmount:      discountAmount,
		TaxAmount:           taxAmount,
		ServiceChargeAmount: serviceCharge,
		TotalAmount:         totalAmount,
		PaidAmount:          0,
		PromoID:             appliedPromoID,
		CreatedByUserID:     creatorID,
		CashShiftID:         req.CashShiftID,
		Notes:               strings.TrimSpace(req.Notes),
		Items:               orderItems,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.orderRepo.CreateOrder(ctx, order, orderItems); err != nil {
		return nil, err
	}

	// Update status meja jika Dine-In
	if order.TableID != nil && *order.TableID != "" {
		_ = s.tableRepo.UpdateTableStatus(ctx, *order.TableID, "occupied")
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     creatorID,
		Action:     "order.create",
		EntityType: "order",
		EntityID:   &order.ID,
		Details:    fmt.Sprintf(`{"number":"%s","total":%d,"type":"%s"}`, order.OrderNumber, order.TotalAmount, order.OrderType),
		IPAddress:  ip,
	})

	return order, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	return s.orderRepo.GetOrderByID(ctx, id)
}

func (s *OrderService) ListOrders(ctx context.Context, status, paymentStatus, orderType, dateStr string, page, perPage int) ([]*models.Order, *models.PaginationMeta, error) {
	orders, total, err := s.orderRepo.ListOrders(ctx, status, paymentStatus, orderType, dateStr, page, perPage)
	if err != nil {
		return nil, nil, err
	}
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	meta := &models.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: totalPages,
	}
	return orders, meta, nil
}

func (s *OrderService) ListKitchenOrders(ctx context.Context) ([]*models.Order, error) {
	return s.orderRepo.ListKitchenOrders(ctx)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID, status string, actor *models.User, ip string) error {
	if err := s.orderRepo.UpdateOrderStatus(ctx, orderID, status); err != nil {
		return err
	}

	var actorID *string
	if actor != nil {
		actorID = &actor.ID
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     actorID,
		Action:     "order.status_update",
		EntityType: "order",
		EntityID:   &orderID,
		Details:    fmt.Sprintf(`{"new_status":"%s"}`, status),
		IPAddress:  ip,
	})

	return nil
}

func (s *OrderService) UpdateOrderItemStatus(ctx context.Context, itemID, status string) error {
	return s.orderRepo.UpdateOrderItemStatus(ctx, itemID, status)
}
