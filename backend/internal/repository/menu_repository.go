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
	ErrCategoryNotFound = errors.New("kategori menu tidak ditemukan")
	ErrCategoryHasItems = errors.New("kategori tidak dapat dihapus karena masih memiliki menu terkait")
	ErrMenuItemNotFound = errors.New("item menu tidak ditemukan")
	ErrSlugExists       = errors.New("slug nama menu sudah digunakan, silakan pilih nama lain")
)

// MenuRepository mengelola operasi database untuk kategori, menu, varian, dan modifier.
type MenuRepository struct {
	db *database.DB
}

// NewMenuRepository membuat instance MenuRepository baru.
func NewMenuRepository(db *database.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// --- KATEGORI MENU ---

func (r *MenuRepository) ListCategories(ctx context.Context, onlyActive bool) ([]*models.MenuCategory, error) {
	query := `SELECT id, name, slug, sort_order, is_active, created_at, updated_at FROM menu_categories`
	if onlyActive {
		query += ` WHERE is_active = 1`
	}
	query += ` ORDER BY sort_order ASC, name ASC;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query list kategori gagal: %w", err)
	}
	defer rows.Close()

	var categories []*models.MenuCategory
	for rows.Next() {
		c := &models.MenuCategory{}
		var isActiveInt int
		var createdStr, updatedStr string
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder, &isActiveInt, &createdStr, &updatedStr); err != nil {
			return nil, err
		}
		c.IsActive = isActiveInt == 1
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *MenuRepository) GetCategoryByID(ctx context.Context, id string) (*models.MenuCategory, error) {
	query := `SELECT id, name, slug, sort_order, is_active, created_at, updated_at FROM menu_categories WHERE id = ?;`
	c := &models.MenuCategory{}
	var isActiveInt int
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder, &isActiveInt, &createdStr, &updatedStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	c.IsActive = isActiveInt == 1
	c.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return c, nil
}

func (r *MenuRepository) CreateCategory(ctx context.Context, c *models.MenuCategory) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	isActiveInt := 0
	if c.IsActive {
		isActiveInt = 1
	}

	query := `
		INSERT INTO menu_categories (id, name, slug, sort_order, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Name, c.Slug, c.SortOrder, isActiveInt, nowStr, nowStr)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: menu_categories.slug") {
			return ErrSlugExists
		}
		return fmt.Errorf("create category gagal: %w", err)
	}
	return nil
}

func (r *MenuRepository) UpdateCategory(ctx context.Context, c *models.MenuCategory) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	isActiveInt := 0
	if c.IsActive {
		isActiveInt = 1
	}

	query := `
		UPDATE menu_categories
		SET name = ?, slug = ?, sort_order = ?, is_active = ?, updated_at = ?
		WHERE id = ?;
	`
	res, err := r.db.ExecContext(ctx, query, c.Name, c.Slug, c.SortOrder, isActiveInt, nowStr, c.ID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: menu_categories.slug") {
			return ErrSlugExists
		}
		return fmt.Errorf("update category gagal: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *MenuRepository) DeleteCategory(ctx context.Context, id string) error {
	// Pastikan tidak ada menu item yang memakai kategori ini
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM menu_items WHERE category_id = ? AND deleted_at IS NULL;", id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrCategoryHasItems
	}

	_, err := r.db.ExecContext(ctx, "DELETE FROM menu_categories WHERE id = ?;", id)
	return err
}

// --- MENU ITEMS ---

func (r *MenuRepository) ListMenuItems(ctx context.Context, search, categoryID string, availableOnly bool, page, perPage int) ([]*models.MenuItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var where []string
	var args []interface{}

	where = append(where, "m.deleted_at IS NULL")

	if search != "" {
		where = append(where, "(m.name LIKE ? OR m.description LIKE ?)")
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}
	if categoryID != "" {
		where = append(where, "m.category_id = ?")
		args = append(args, categoryID)
	}
	if availableOnly {
		where = append(where, "m.is_available = 1")
	}

	whereSQL := strings.Join(where, " AND ")

	var totalItems int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM menu_items m WHERE %s;", whereSQL)
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&totalItems); err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(`
		SELECT m.id, m.category_id, c.name, m.name, m.slug, m.description, m.base_price,
		       m.is_available, m.is_best_seller, m.is_featured, m.image_url, m.sort_order, m.created_at, m.updated_at
		FROM menu_items m
		JOIN menu_categories c ON m.category_id = c.id
		WHERE %s
		ORDER BY c.sort_order ASC, m.sort_order ASC, m.name ASC
		LIMIT ? OFFSET ?;
	`, whereSQL)

	queryArgs := append(args, perPage, offset)
	rows, err := r.db.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*models.MenuItem
	for rows.Next() {
		m := &models.MenuItem{}
		var desc, img sql.NullString
		var avail, best, feat int
		var createdStr, updatedStr string

		if err := rows.Scan(&m.ID, &m.CategoryID, &m.CategoryName, &m.Name, &m.Slug, &desc, &m.BasePrice,
			&avail, &best, &feat, &img, &m.SortOrder, &createdStr, &updatedStr); err != nil {
			return nil, 0, err
		}

		if desc.Valid {
			m.Description = desc.String
		}
		if img.Valid {
			m.ImageURL = img.String
		}
		m.IsAvailable = avail == 1
		m.IsBestSeller = best == 1
		m.IsFeatured = feat == 1
		m.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		m.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

		items = append(items, m)
	}

	return items, totalItems, nil
}

func (r *MenuRepository) GetMenuItemByID(ctx context.Context, id string) (*models.MenuItem, error) {
	query := `
		SELECT m.id, m.category_id, c.name, m.name, m.slug, m.description, m.base_price,
		       m.is_available, m.is_best_seller, m.is_featured, m.image_url, m.sort_order, m.created_at, m.updated_at
		FROM menu_items m
		JOIN menu_categories c ON m.category_id = c.id
		WHERE m.id = ? AND m.deleted_at IS NULL;
	`
	m := &models.MenuItem{}
	var desc, img sql.NullString
	var avail, best, feat int
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.CategoryID, &m.CategoryName, &m.Name, &m.Slug, &desc,
		&m.BasePrice, &avail, &best, &feat, &img, &m.SortOrder, &createdStr, &updatedStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMenuItemNotFound
		}
		return nil, err
	}

	if desc.Valid {
		m.Description = desc.String
	}
	if img.Valid {
		m.ImageURL = img.String
	}
	m.IsAvailable = avail == 1
	m.IsBestSeller = best == 1
	m.IsFeatured = feat == 1
	m.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	m.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	// Ambil variants
	m.Variants, _ = r.GetVariantsByMenuItemID(ctx, m.ID)

	// Ambil modifiers
	m.Modifiers, _ = r.GetModifierGroupsByMenuItemID(ctx, m.ID)

	return m, nil
}

func (r *MenuRepository) CreateMenuItem(ctx context.Context, m *models.MenuItem, variants []models.CreateVariant, modifierGroupIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)
	avail := 0
	if m.IsAvailable {
		avail = 1
	}
	best := 0
	if m.IsBestSeller {
		best = 1
	}
	feat := 0
	if m.IsFeatured {
		feat = 1
	}

	insertSQL := `
		INSERT INTO menu_items (id, category_id, name, slug, description, base_price, is_available, is_best_seller, is_featured, image_url, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertSQL, m.ID, m.CategoryID, m.Name, m.Slug, m.Description, m.BasePrice, avail, best, feat, m.ImageURL, m.SortOrder, nowStr, nowStr)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: menu_items.slug") {
			return ErrSlugExists
		}
		return err
	}

	// Insert variants
	variantSQL := `INSERT INTO menu_variants (id, menu_item_id, name, price_adjustment, is_available, sort_order, created_at) VALUES (?, ?, ?, ?, ?, ?, ?);`
	for idx, v := range variants {
		vID := fmt.Sprintf("var-%s-%d", m.ID, idx+1)
		vAvail := 0
		if v.IsAvailable {
			vAvail = 1
		}
		if _, err := tx.ExecContext(ctx, variantSQL, vID, m.ID, v.Name, v.PriceAdjustment, vAvail, v.SortOrder, nowStr); err != nil {
			return err
		}
	}

	// Link modifier groups
	linkSQL := `INSERT INTO menu_item_modifier_groups (menu_item_id, modifier_group_id) VALUES (?, ?);`
	for _, mgID := range modifierGroupIDs {
		if _, err := tx.ExecContext(ctx, linkSQL, m.ID, mgID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MenuRepository) UpdateMenuItem(ctx context.Context, m *models.MenuItem) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	avail := 0
	if m.IsAvailable {
		avail = 1
	}
	best := 0
	if m.IsBestSeller {
		best = 1
	}
	feat := 0
	if m.IsFeatured {
		feat = 1
	}

	query := `
		UPDATE menu_items
		SET category_id = ?, name = ?, slug = ?, description = ?, base_price = ?,
		    is_available = ?, is_best_seller = ?, is_featured = ?, image_url = ?, sort_order = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL;
	`
	res, err := r.db.ExecContext(ctx, query, m.CategoryID, m.Name, m.Slug, m.Description, m.BasePrice, avail, best, feat, m.ImageURL, m.SortOrder, nowStr, m.ID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: menu_items.slug") {
			return ErrSlugExists
		}
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrMenuItemNotFound
	}
	return nil
}

func (r *MenuRepository) ToggleAvailability(ctx context.Context, id string, isAvailable bool) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	val := 0
	if isAvailable {
		val = 1
	}
	res, err := r.db.ExecContext(ctx, "UPDATE menu_items SET is_available = ?, updated_at = ? WHERE id = ?;", val, nowStr, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrMenuItemNotFound
	}
	return nil
}

func (r *MenuRepository) SoftDeleteMenuItem(ctx context.Context, id string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, "UPDATE menu_items SET deleted_at = ?, is_available = 0, updated_at = ? WHERE id = ?;", nowStr, nowStr, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrMenuItemNotFound
	}
	return nil
}

// Helpers Variants & Modifiers

func (r *MenuRepository) GetVariantsByMenuItemID(ctx context.Context, itemID string) ([]*models.MenuVariant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, menu_item_id, name, price_adjustment, is_available, sort_order
		FROM menu_variants
		WHERE menu_item_id = ?
		ORDER BY sort_order ASC;
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.MenuVariant
	for rows.Next() {
		v := &models.MenuVariant{}
		var avail int
		if err := rows.Scan(&v.ID, &v.MenuItemID, &v.Name, &v.PriceAdjustment, &avail, &v.SortOrder); err != nil {
			return nil, err
		}
		v.IsAvailable = avail == 1
		list = append(list, v)
	}
	return list, nil
}

func (r *MenuRepository) GetModifierGroupsByMenuItemID(ctx context.Context, itemID string) ([]*models.ModifierGroup, error) {
	query := `
		SELECT mg.id, mg.name, mg.is_required, mg.min_select, mg.max_select
		FROM modifier_groups mg
		JOIN menu_item_modifier_groups mmg ON mg.id = mmg.modifier_group_id
		WHERE mmg.menu_item_id = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*models.ModifierGroup
	for rows.Next() {
		g := &models.ModifierGroup{}
		var req int
		if err := rows.Scan(&g.ID, &g.Name, &req, &g.MinSelect, &g.MaxSelect); err != nil {
			return nil, err
		}
		g.IsRequired = req == 1

		// Fetch options
		optRows, err := r.db.QueryContext(ctx, `
			SELECT id, modifier_group_id, name, additional_price, is_available, sort_order
			FROM modifier_options
			WHERE modifier_group_id = ?
			ORDER BY sort_order ASC;
		`, g.ID)
		if err == nil {
			for optRows.Next() {
				opt := &models.ModifierOption{}
				var optAvail int
				if err := optRows.Scan(&opt.ID, &opt.ModifierGroupID, &opt.Name, &opt.AdditionalPrice, &optAvail, &opt.SortOrder); err == nil {
					opt.IsAvailable = optAvail == 1
					g.Options = append(g.Options, opt)
				}
			}
			optRows.Close()
		}

		groups = append(groups, g)
	}
	return groups, nil
}
