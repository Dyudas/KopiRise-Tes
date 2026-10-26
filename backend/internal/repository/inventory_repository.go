package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/models"
)

var (
	ErrIngredientNotFound = errors.New("bahan baku tidak ditemukan")
	ErrSupplierNotFound   = errors.New("supplier tidak ditemukan")
	ErrRecipeNotFound     = errors.New("resep menu tidak ditemukan")
)

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// ==========================================
// 1. BAHAN BAKU (INGREDIENTS)
// ==========================================

func (r *InventoryRepository) ListIngredients(ctx context.Context) ([]*models.Ingredient, error) {
	query := `
		SELECT id, name, unit, current_stock, minimum_stock, cost_per_unit, is_active, created_at, updated_at
		FROM ingredients
		ORDER BY name ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list ingredients gagal: %w", err)
	}
	defer rows.Close()

	var list []*models.Ingredient
	for rows.Next() {
		var i models.Ingredient
		var isActiveInt int
		var createdStr, updatedStr string
		if err := rows.Scan(
			&i.ID, &i.Name, &i.Unit, &i.CurrentStock, &i.MinimumStock,
			&i.CostPerUnit, &isActiveInt, &createdStr, &updatedStr,
		); err != nil {
			return nil, err
		}
		i.IsActive = isActiveInt == 1
		i.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		i.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
		list = append(list, &i)
	}
	return list, nil
}

func (r *InventoryRepository) ListLowStockIngredients(ctx context.Context) ([]*models.Ingredient, error) {
	query := `
		SELECT id, name, unit, current_stock, minimum_stock, cost_per_unit, is_active, created_at, updated_at
		FROM ingredients
		WHERE current_stock <= minimum_stock AND is_active = 1
		ORDER BY (current_stock / CASE WHEN minimum_stock = 0 THEN 1 ELSE minimum_stock END) ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Ingredient
	for rows.Next() {
		var i models.Ingredient
		var isActiveInt int
		var createdStr, updatedStr string
		if err := rows.Scan(
			&i.ID, &i.Name, &i.Unit, &i.CurrentStock, &i.MinimumStock,
			&i.CostPerUnit, &isActiveInt, &createdStr, &updatedStr,
		); err != nil {
			return nil, err
		}
		i.IsActive = isActiveInt == 1
		i.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		i.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
		list = append(list, &i)
	}
	return list, nil
}

func (r *InventoryRepository) GetIngredientByID(ctx context.Context, id string) (*models.Ingredient, error) {
	query := `
		SELECT id, name, unit, current_stock, minimum_stock, cost_per_unit, is_active, created_at, updated_at
		FROM ingredients
		WHERE id = ?;
	`
	var i models.Ingredient
	var isActiveInt int
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&i.ID, &i.Name, &i.Unit, &i.CurrentStock, &i.MinimumStock,
		&i.CostPerUnit, &isActiveInt, &createdStr, &updatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrIngredientNotFound
		}
		return nil, err
	}
	i.IsActive = isActiveInt == 1
	i.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	i.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return &i, nil
}

func (r *InventoryRepository) CreateIngredient(ctx context.Context, i *models.Ingredient) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	isActiveInt := 1
	if !i.IsActive {
		isActiveInt = 0
	}

	query := `
		INSERT INTO ingredients (id, name, unit, current_stock, minimum_stock, cost_per_unit, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		i.ID, i.Name, i.Unit, i.CurrentStock, i.MinimumStock, i.CostPerUnit, isActiveInt, nowStr, nowStr,
	)
	return err
}

func (r *InventoryRepository) UpdateIngredient(ctx context.Context, i *models.Ingredient) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	isActiveInt := 1
	if !i.IsActive {
		isActiveInt = 0
	}

	query := `
		UPDATE ingredients
		SET name = ?, unit = ?, minimum_stock = ?, is_active = ?, updated_at = ?
		WHERE id = ?;
	`
	res, err := r.db.ExecContext(ctx, query, i.Name, i.Unit, i.MinimumStock, isActiveInt, nowStr, i.ID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrIngredientNotFound
	}
	return nil
}

// ==========================================
// 2. SUPPLIERS
// ==========================================

func (r *InventoryRepository) ListSuppliers(ctx context.Context) ([]*models.Supplier, error) {
	query := `
		SELECT id, name, COALESCE(phone, ''), COALESCE(email, ''), COALESCE(address, ''), COALESCE(contact_person, ''), created_at
		FROM suppliers
		ORDER BY name ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Supplier
	for rows.Next() {
		var s models.Supplier
		var createdStr string
		if err := rows.Scan(&s.ID, &s.Name, &s.Phone, &s.Email, &s.Address, &s.ContactPerson, &createdStr); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		list = append(list, &s)
	}
	return list, nil
}

func (r *InventoryRepository) CreateSupplier(ctx context.Context, s *models.Supplier) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO suppliers (id, name, phone, email, address, contact_person, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.Name, s.Phone, s.Email, s.Address, s.ContactPerson, nowStr)
	return err
}

// ==========================================
// 3. PENGADAAN & PEMBELIAN STOK (PURCHASES)
// ==========================================

// CreatePurchase mencatat pembelian bahan baku, menghitung biaya rata-rata tertimbang, memperbarui stok, dan memotong saldo rekening finansial secara atomik.
func (r *InventoryRepository) CreatePurchase(ctx context.Context, p *models.IngredientPurchase, items []*models.IngredientPurchaseItem) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// 1. Insert Purchase Record
	insertPurchaseSQL := `
		INSERT INTO ingredient_purchases (
			id, purchase_number, supplier_id, financial_account_id, purchase_date,
			total_amount, expense_id, created_by_user_id, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertPurchaseSQL,
		p.ID, p.PurchaseNumber, p.SupplierID, p.FinancialAccountID, p.PurchaseDate,
		p.TotalAmount, p.ExpenseID, p.CreatedByUserID, p.Notes, nowStr,
	)
	if err != nil {
		return fmt.Errorf("insert purchase gagal: %w", err)
	}

	// 2. Loop Items: Hitung Weighted Average Cost, Update Stok, Catat Kartu Stok (Movements)
	insertItemSQL := `
		INSERT INTO ingredient_purchase_items (id, purchase_id, ingredient_id, quantity, unit_cost, subtotal)
		VALUES (?, ?, ?, ?, ?, ?);
	`
	insertMovementSQL := `
		INSERT INTO inventory_movements (
			id, ingredient_id, movement_type, quantity, previous_stock, new_stock,
			reference_type, reference_id, reason, performed_by_user_id, created_at
		) VALUES (?, ?, 'purchase', ?, ?, ?, 'purchase', ?, 'Pembelian Stok Bahan', ?, ?);
	`

	for _, item := range items {
		// Insert item
		_, err = tx.ExecContext(ctx, insertItemSQL, item.ID, p.ID, item.IngredientID, item.Quantity, item.UnitCost, item.Subtotal)
		if err != nil {
			return fmt.Errorf("insert purchase item gagal: %w", err)
		}

		// Ambil stok dan cost lama
		var oldStock float64
		var oldCost int64
		err = tx.QueryRowContext(ctx, "SELECT current_stock, cost_per_unit FROM ingredients WHERE id = ?;", item.IngredientID).Scan(&oldStock, &oldCost)
		if err != nil {
			return fmt.Errorf("ambil data bahan %s gagal: %w", item.IngredientID, err)
		}

		newStock := oldStock + item.Quantity
		// Formula Biaya Rata-Rata Tertimbang (Weighted Average Cost):
		// new_cost = ((old_stock * old_cost) + (purchased_qty * unit_cost)) / new_stock
		var newCost int64
		if newStock > 0 {
			totalValuation := (oldStock * float64(oldCost)) + (item.Quantity * float64(item.UnitCost))
			newCost = int64(totalValuation / newStock)
		} else {
			newCost = item.UnitCost
		}

		// Update ingredients
		_, err = tx.ExecContext(ctx, "UPDATE ingredients SET current_stock = ?, cost_per_unit = ?, updated_at = ? WHERE id = ?;", newStock, newCost, nowStr, item.IngredientID)
		if err != nil {
			return fmt.Errorf("update stock & cost bahan gagal: %w", err)
		}

		// Catat riwayat kartu stok
		movID := "mov-" + uuid.New().String()[:8]
		_, err = tx.ExecContext(ctx, insertMovementSQL, movID, item.IngredientID, item.Quantity, oldStock, newStock, p.ID, p.CreatedByUserID, nowStr)
		if err != nil {
			return fmt.Errorf("insert inventory movement gagal: %w", err)
		}
	}

	// 3. Potong saldo akun keuangan yang digunakan untuk membayar pembelian
	_, err = tx.ExecContext(ctx, "UPDATE financial_accounts SET current_balance = current_balance - ?, updated_at = ? WHERE id = ?;", p.TotalAmount, nowStr, p.FinancialAccountID)
	if err != nil {
		return fmt.Errorf("potong saldo akun pembayaran pembelian gagal: %w", err)
	}

	return tx.Commit()
}

func (r *InventoryRepository) ListPurchases(ctx context.Context, page, perPage int) ([]*models.IngredientPurchase, *models.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ingredient_purchases;").Scan(&total)

	query := `
		SELECT ip.id, ip.purchase_number, ip.supplier_id, COALESCE(s.name, '-'),
		       ip.financial_account_id, fa.name, ip.purchase_date, ip.total_amount,
		       ip.created_by_user_id, u.name, COALESCE(ip.notes, ''), ip.created_at
		FROM ingredient_purchases ip
		LEFT JOIN suppliers s ON ip.supplier_id = s.id
		JOIN financial_accounts fa ON ip.financial_account_id = fa.id
		JOIN users u ON ip.created_by_user_id = u.id
		ORDER BY ip.purchase_date DESC, ip.created_at DESC
		LIMIT ? OFFSET ?;
	`
	rows, err := r.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var purchases []*models.IngredientPurchase
	for rows.Next() {
		var p models.IngredientPurchase
		var suppName sql.NullString
		var createdStr string
		if err := rows.Scan(
			&p.ID, &p.PurchaseNumber, &p.SupplierID, &suppName,
			&p.FinancialAccountID, &p.AccountName, &p.PurchaseDate, &p.TotalAmount,
			&p.CreatedByUserID, &p.CreatedByUserName, &p.Notes, &createdStr,
		); err != nil {
			return nil, nil, err
		}
		if suppName.Valid {
			p.SupplierName = &suppName.String
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		purchases = append(purchases, &p)
	}

	totalPages := (total + perPage - 1) / perPage
	meta := &models.PaginationMeta{
		CurrentPage: page,
		PerPage:     perPage,
		TotalItems:  total,
		TotalPages:  totalPages,
	}
	return purchases, meta, nil
}

// ==========================================
// 4. RESEP (BILL OF MATERIALS) & HPP
// ==========================================

func (r *InventoryRepository) GetRecipeByMenuItemID(ctx context.Context, menuItemID string) (*models.Recipe, error) {
	query := `
		SELECT r.id, r.menu_item_id, m.name, m.base_price, COALESCE(r.notes, ''), r.created_at, r.updated_at
		FROM recipes r
		JOIN menu_items m ON r.menu_item_id = m.id
		WHERE r.menu_item_id = ?;
	`
	var rec models.Recipe
	var createdStr, updatedStr string
	err := r.db.QueryRowContext(ctx, query, menuItemID).Scan(
		&rec.ID, &rec.MenuItemID, &rec.MenuItemName, &rec.SellingPrice, &rec.Notes, &createdStr, &updatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecipeNotFound
		}
		return nil, err
	}
	rec.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	rec.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	// Fetch recipe items and calculate HPP
	itemQuery := `
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ing.name, ing.unit, ri.quantity, ing.cost_per_unit
		FROM recipe_items ri
		JOIN ingredients ing ON ri.ingredient_id = ing.id
		WHERE ri.recipe_id = ?;
	`
	rows, err := r.db.QueryContext(ctx, itemQuery, rec.ID)
	if err == nil {
		defer rows.Close()
		var totalHPP int64
		for rows.Next() {
			var ri models.RecipeItem
			if err := rows.Scan(&ri.ID, &ri.RecipeID, &ri.IngredientID, &ri.IngredientName, &ri.Unit, &ri.Quantity, &ri.CostPerUnit); err == nil {
				ri.SubtotalCost = int64(ri.Quantity * float64(ri.CostPerUnit))
				totalHPP += ri.SubtotalCost
				rec.Items = append(rec.Items, &ri)
			}
		}
		rec.TotalHPP = totalHPP
		rec.MarginRp = rec.SellingPrice - totalHPP
		if rec.SellingPrice > 0 {
			rec.MarginPct = (float64(rec.MarginRp) / float64(rec.SellingPrice)) * 100.0
		}
	}

	return &rec, nil
}

func (r *InventoryRepository) CreateOrUpdateRecipe(ctx context.Context, rec *models.Recipe, items []*models.RecipeItem) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// Periksa apakah resep untuk menu ini sudah ada
	var existingRecipeID string
	err = tx.QueryRowContext(ctx, "SELECT id FROM recipes WHERE menu_item_id = ?;", rec.MenuItemID).Scan(&existingRecipeID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if existingRecipeID != "" {
		// Update existing recipe & delete old items
		rec.ID = existingRecipeID
		_, err = tx.ExecContext(ctx, "UPDATE recipes SET notes = ?, updated_at = ? WHERE id = ?;", rec.Notes, nowStr, rec.ID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "DELETE FROM recipe_items WHERE recipe_id = ?;", rec.ID)
		if err != nil {
			return err
		}
	} else {
		// Insert new recipe
		_, err = tx.ExecContext(ctx, "INSERT INTO recipes (id, menu_item_id, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?);",
			rec.ID, rec.MenuItemID, rec.Notes, nowStr, nowStr,
		)
		if err != nil {
			return err
		}
	}

	// Insert items
	insertItemSQL := `INSERT INTO recipe_items (id, recipe_id, ingredient_id, quantity, created_at) VALUES (?, ?, ?, ?, ?);`
	for _, item := range items {
		_, err = tx.ExecContext(ctx, insertItemSQL, item.ID, rec.ID, item.IngredientID, item.Quantity, nowStr)
		if err != nil {
			return fmt.Errorf("insert recipe item gagal: %w", err)
		}
	}

	return tx.Commit()
}

// DeductStockForOrder memotong stok bahan baku secara otomatis ketika pesanan selesai dimasak/dijual.
func (r *InventoryRepository) DeductStockForOrder(ctx context.Context, orderID string, performedByUserID *string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// Ambil semua item dalam order
	query := `SELECT menu_item_id, quantity FROM order_items WHERE order_id = ? AND menu_item_id IS NOT NULL;`
	rows, err := tx.QueryContext(ctx, query, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type OrderLine struct {
		MenuItemID string
		Qty        int
	}
	var lines []OrderLine
	for rows.Next() {
		var l OrderLine
		if err := rows.Scan(&l.MenuItemID, &l.Qty); err == nil {
			lines = append(lines, l)
		}
	}

	// Untuk setiap menu item, cari bahan baku dalam resep
	recipeItemQuery := `
		SELECT ri.ingredient_id, ri.quantity, ing.current_stock
		FROM recipe_items ri
		JOIN recipes r ON ri.recipe_id = r.id
		JOIN ingredients ing ON ri.ingredient_id = ing.id
		WHERE r.menu_item_id = ?;
	`
	insertMovementSQL := `
		INSERT INTO inventory_movements (
			id, ingredient_id, movement_type, quantity, previous_stock, new_stock,
			reference_type, reference_id, reason, performed_by_user_id, created_at
		) VALUES (?, ?, 'sale_consumption', ?, ?, ?, 'order', ?, 'Pengurangan Stok Penjualan', ?, ?);
	`

	for _, line := range lines {
		rRows, err := tx.QueryContext(ctx, recipeItemQuery, line.MenuItemID)
		if err != nil {
			continue
		}

		type Deduction struct {
			IngID    string
			Deduct   float64
			Prev     float64
		}
		var deductions []Deduction
		for rRows.Next() {
			var ingID string
			var perPortionQty, prevStock float64
			if err := rRows.Scan(&ingID, &perPortionQty, &prevStock); err == nil {
				totalDeduct := perPortionQty * float64(line.Qty)
				deductions = append(deductions, Deduction{IngID: ingID, Deduct: totalDeduct, Prev: prevStock})
			}
		}
		rRows.Close()

		for _, d := range deductions {
			newStock := d.Prev - d.Deduct
			if newStock < 0 {
				newStock = 0 // Cegah stok negatif pada kartu persediaan
			}

			_, err = tx.ExecContext(ctx, "UPDATE ingredients SET current_stock = ?, updated_at = ? WHERE id = ?;", newStock, nowStr, d.IngID)
			if err != nil {
				return err
			}

			movID := "mov-" + uuid.New().String()[:8]
			_, _ = tx.ExecContext(ctx, insertMovementSQL, movID, d.IngID, d.Deduct, d.Prev, newStock, orderID, performedByUserID, nowStr)
		}
	}

	return tx.Commit()
}
