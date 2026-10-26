package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"kopirise/backend/internal/database"
	"kopirise/backend/internal/models"
)

var (
	ErrOrderNotFound          = errors.New("pesanan tidak ditemukan")
	ErrDuplicateIdempotency   = errors.New("permintaan dengan idempotency key ini sudah pernah diproses")
	ErrInvalidOrderStatus     = errors.New("transisi status pesanan tidak valid")
	ErrPromoNotFoundOrExpired = errors.New("kode promo tidak ditemukan atau sudah kedaluwarsa")
)

type OrderRepository struct {
	db *database.DB
}

func NewOrderRepository(db *database.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder menyimpan order dan snapshots item dalam satu transaksi SQL atomik.
func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	insertOrderSQL := `
		INSERT INTO orders (
			id, order_number, idempotency_key, order_type, table_id,
			customer_name, customer_phone, status, payment_status,
			subtotal, discount_amount, tax_amount, service_charge_amount,
			total_amount, paid_amount, promo_id, created_by_user_id,
			cash_shift_id, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertOrderSQL,
		order.ID,
		order.OrderNumber,
		order.IdempotencyKey,
		order.OrderType,
		order.TableID,
		order.CustomerName,
		order.CustomerPhone,
		order.Status,
		order.PaymentStatus,
		order.Subtotal,
		order.DiscountAmount,
		order.TaxAmount,
		order.ServiceChargeAmount,
		order.TotalAmount,
		order.PaidAmount,
		order.PromoID,
		order.CreatedByUserID,
		order.CashShiftID,
		order.Notes,
		nowStr,
		nowStr,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: orders.idempotency_key") {
			return ErrDuplicateIdempotency
		}
		return fmt.Errorf("insert order gagal: %w", err)
	}

	insertItemSQL := `
		INSERT INTO order_items (
			id, order_id, menu_item_id, menu_variant_id,
			item_name_snapshot, variant_name_snapshot, unit_base_price_snapshot,
			quantity, subtotal, notes, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	insertModSQL := `
		INSERT INTO order_item_modifiers (
			id, order_item_id, modifier_option_id,
			group_name_snapshot, option_name_snapshot, additional_price_snapshot
		) VALUES (?, ?, ?, ?, ?, ?);
	`

	for _, itm := range items {
		_, err = tx.ExecContext(ctx, insertItemSQL,
			itm.ID,
			order.ID,
			itm.MenuItemID,
			itm.MenuVariantID,
			itm.ItemNameSnapshot,
			itm.VariantNameSnapshot,
			itm.UnitBasePriceSnapshot,
			itm.Quantity,
			itm.Subtotal,
			itm.Notes,
			itm.Status,
			nowStr,
		)
		if err != nil {
			return fmt.Errorf("insert order item gagal: %w", err)
		}

		for _, mod := range itm.Modifiers {
			_, err = tx.ExecContext(ctx, insertModSQL,
				mod.ID,
				itm.ID,
				mod.ModifierOptionID,
				mod.GroupNameSnapshot,
				mod.OptionNameSnapshot,
				mod.AdditionalPriceSnapshot,
			)
			if err != nil {
				return fmt.Errorf("insert modifier snapshot gagal: %w", err)
			}
		}
	}

	// Update promo usage if applied
	if order.PromoID != nil {
		_, _ = tx.ExecContext(ctx, "UPDATE promos SET used_count = used_count + 1 WHERE id = ?;", *order.PromoID)
	}

	return tx.Commit()
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	query := `
		SELECT o.id, o.order_number, o.idempotency_key, o.order_type, o.table_id,
		       COALESCE(t.table_number, '') AS table_number,
		       o.customer_name, o.customer_phone, o.status, o.payment_status,
		       o.subtotal, o.discount_amount, o.tax_amount, o.service_charge_amount,
		       o.total_amount, o.paid_amount, o.promo_id, COALESCE(p.code, '') AS promo_code,
		       o.created_by_user_id, o.cash_shift_id, o.notes, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN cafe_tables t ON o.table_id = t.id
		LEFT JOIN promos p ON o.promo_id = p.id
		WHERE o.id = ?;
	`
	ord := &models.Order{}
	var idemKey, tblID, custName, custPhone, promoID, creatorID, shiftID, notes sql.NullString
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ord.ID, &ord.OrderNumber, &idemKey, &ord.OrderType, &tblID, &ord.TableNumber,
		&custName, &custPhone, &ord.Status, &ord.PaymentStatus,
		&ord.Subtotal, &ord.DiscountAmount, &ord.TaxAmount, &ord.ServiceChargeAmount,
		&ord.TotalAmount, &ord.PaidAmount, &promoID, &ord.PromoCode,
		&creatorID, &shiftID, &notes, &createdStr, &updatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	if idemKey.Valid {
		ord.IdempotencyKey = &idemKey.String
	}
	if tblID.Valid {
		ord.TableID = &tblID.String
	}
	if custName.Valid {
		ord.CustomerName = custName.String
	}
	if custPhone.Valid {
		ord.CustomerPhone = custPhone.String
	}
	if promoID.Valid {
		ord.PromoID = &promoID.String
	}
	if creatorID.Valid {
		ord.CreatedByUserID = &creatorID.String
	}
	if shiftID.Valid {
		ord.CashShiftID = &shiftID.String
	}
	if notes.Valid {
		ord.Notes = notes.String
	}
	ord.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	ord.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	// Fetch items
	items, err := r.getOrderItems(ctx, ord.ID)
	if err == nil {
		ord.Items = items
	}

	return ord, nil
}

func (r *OrderRepository) GetOrderByIdempotencyKey(ctx context.Context, key string) (*models.Order, error) {
	var id string
	err := r.db.QueryRowContext(ctx, "SELECT id FROM orders WHERE idempotency_key = ?;", key).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return r.GetOrderByID(ctx, id)
}

func (r *OrderRepository) ListOrders(ctx context.Context, status, paymentStatus, orderType, dateStr string, page, perPage int) ([]*models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var where []string
	var args []interface{}

	if status != "" {
		where = append(where, "o.status = ?")
		args = append(args, status)
	}
	if paymentStatus != "" {
		where = append(where, "o.payment_status = ?")
		args = append(args, paymentStatus)
	}
	if orderType != "" {
		where = append(where, "o.order_type = ?")
		args = append(args, orderType)
	}
	if dateStr != "" {
		where = append(where, "date(o.created_at) = ?")
		args = append(args, dateStr)
	}

	whereSQL := "WHERE 1=1"
	if len(where) > 0 {
		whereSQL += " AND " + strings.Join(where, " AND ")
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM orders o %s;", whereSQL)
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(`
		SELECT o.id, o.order_number, o.order_type, o.table_id, COALESCE(t.table_number, '') AS table_number,
		       o.customer_name, o.customer_phone, o.status, o.payment_status,
		       o.subtotal, o.discount_amount, o.tax_amount, o.service_charge_amount,
		       o.total_amount, o.paid_amount, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN cafe_tables t ON o.table_id = t.id
		%s
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?;
	`, whereSQL)

	queryArgs := append(args, perPage, offset)
	rows, err := r.db.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		ord := &models.Order{}
		var tblID, custName, custPhone sql.NullString
		var createdStr, updatedStr string

		if err := rows.Scan(
			&ord.ID, &ord.OrderNumber, &ord.OrderType, &tblID, &ord.TableNumber,
			&custName, &custPhone, &ord.Status, &ord.PaymentStatus,
			&ord.Subtotal, &ord.DiscountAmount, &ord.TaxAmount, &ord.ServiceChargeAmount,
			&ord.TotalAmount, &ord.PaidAmount, &createdStr, &updatedStr,
		); err != nil {
			return nil, 0, err
		}

		if tblID.Valid {
			ord.TableID = &tblID.String
		}
		if custName.Valid {
			ord.CustomerName = custName.String
		}
		if custPhone.Valid {
			ord.CustomerPhone = custPhone.String
		}
		ord.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		ord.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

		orders = append(orders, ord)
	}

	return orders, total, nil
}

// ListKitchenOrders mengambil seluruh pesanan antrean barista/dapur (status confirmed & preparing).
func (r *OrderRepository) ListKitchenOrders(ctx context.Context) ([]*models.Order, error) {
	query := `
		SELECT o.id, o.order_number, o.order_type, o.table_id, COALESCE(t.table_number, 'Takeaway') AS table_number,
		       COALESCE(o.customer_name, 'Tamu') AS customer_name, o.status, o.notes, o.created_at
		FROM orders o
		LEFT JOIN cafe_tables t ON o.table_id = t.id
		WHERE o.status IN ('confirmed', 'preparing')
		ORDER BY o.created_at ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Order
	for rows.Next() {
		o := &models.Order{}
		var tblID, notes sql.NullString
		var createdStr string
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.OrderType, &tblID, &o.TableNumber, &o.CustomerName, &o.Status, &notes, &createdStr); err != nil {
			return nil, err
		}
		if tblID.Valid {
			o.TableID = &tblID.String
		}
		if notes.Valid {
			o.Notes = notes.String
		}
		o.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)

		// Items
		o.Items, _ = r.getOrderItems(ctx, o.ID)
		list = append(list, o)
	}
	return list, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, "UPDATE orders SET status = ?, updated_at = ? WHERE id = ?;", status, nowStr, orderID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) UpdateOrderItemStatus(ctx context.Context, itemID, status string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE order_items SET status = ? WHERE id = ?;", status, itemID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("item pesanan tidak ditemukan")
	}
	return nil
}

func (r *OrderRepository) getOrderItems(ctx context.Context, orderID string) ([]*models.OrderItem, error) {
	query := `
		SELECT id, order_id, menu_item_id, menu_variant_id,
		       item_name_snapshot, COALESCE(variant_name_snapshot, '') AS variant_name,
		       unit_base_price_snapshot, quantity, subtotal, COALESCE(notes, '') AS notes,
		       status, created_at
		FROM order_items
		WHERE order_id = ?
		ORDER BY id ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.OrderItem
	for rows.Next() {
		it := &models.OrderItem{}
		var mID, vID sql.NullString
		var createdStr string
		if err := rows.Scan(
			&it.ID, &it.OrderID, &mID, &vID,
			&it.ItemNameSnapshot, &it.VariantNameSnapshot,
			&it.UnitBasePriceSnapshot, &it.Quantity, &it.Subtotal,
			&it.Notes, &it.Status, &createdStr,
		); err != nil {
			return nil, err
		}
		if mID.Valid {
			it.MenuItemID = &mID.String
		}
		if vID.Valid {
			it.MenuVariantID = &vID.String
		}
		it.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)

		// Modifiers
		it.Modifiers, _ = r.getOrderItemModifiers(ctx, it.ID)
		items = append(items, it)
	}
	return items, nil
}

func (r *OrderRepository) getOrderItemModifiers(ctx context.Context, itemID string) ([]*models.OrderItemModifier, error) {
	query := `
		SELECT id, order_item_id, modifier_option_id, group_name_snapshot, option_name_snapshot, additional_price_snapshot
		FROM order_item_modifiers
		WHERE order_item_id = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mods []*models.OrderItemModifier
	for rows.Next() {
		m := &models.OrderItemModifier{}
		var optID sql.NullString
		if err := rows.Scan(&m.ID, &m.OrderItemID, &optID, &m.GroupNameSnapshot, &m.OptionNameSnapshot, &m.AdditionalPriceSnapshot); err != nil {
			return nil, err
		}
		if optID.Valid {
			m.ModifierOptionID = &optID.String
		}
		mods = append(mods, m)
	}
	return mods, nil
}
