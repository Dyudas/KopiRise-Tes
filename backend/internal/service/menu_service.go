package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
)

// MenuService mengelola logika bisnis untuk kategori menu, item menu, varian, dan modifier.
type MenuService struct {
	menuRepo *repository.MenuRepository
	audit    *repository.AuditRepository
}

// NewMenuService membuat instance MenuService baru.
func NewMenuService(menuRepo *repository.MenuRepository, audit *repository.AuditRepository) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
		audit:    audit,
	}
}

// --- KATEGORI ---

func (s *MenuService) ListCategories(ctx context.Context, onlyActive bool) ([]*models.MenuCategory, error) {
	return s.menuRepo.ListCategories(ctx, onlyActive)
}

func (s *MenuService) CreateCategory(ctx context.Context, req *models.CreateCategoryRequest, actor *models.User, ip string) (*models.MenuCategory, error) {
	slug := generateSlug(req.Name)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	c := &models.MenuCategory{
		ID:        "cat-" + uuid.NewString()[:8],
		Name:      strings.TrimSpace(req.Name),
		Slug:      slug,
		SortOrder: req.SortOrder,
		IsActive:  isActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.menuRepo.CreateCategory(ctx, c); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "menu.category_create",
		EntityType: "menu_category",
		EntityID:   &c.ID,
		Details:    fmt.Sprintf(`{"name":"%s"}`, c.Name),
		IPAddress:  ip,
	})

	return c, nil
}

func (s *MenuService) UpdateCategory(ctx context.Context, id string, req *models.UpdateCategoryRequest, actor *models.User, ip string) (*models.MenuCategory, error) {
	existing, err := s.menuRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Name = strings.TrimSpace(req.Name)
	existing.Slug = generateSlug(req.Name)
	existing.SortOrder = req.SortOrder
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := s.menuRepo.UpdateCategory(ctx, existing); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "menu.category_update",
		EntityType: "menu_category",
		EntityID:   &existing.ID,
		Details:    fmt.Sprintf(`{"name":"%s"}`, existing.Name),
		IPAddress:  ip,
	})

	return existing, nil
}

func (s *MenuService) DeleteCategory(ctx context.Context, id string, actor *models.User, ip string) error {
	if err := s.menuRepo.DeleteCategory(ctx, id); err != nil {
		return err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "menu.category_delete",
		EntityType: "menu_category",
		EntityID:   &id,
		IPAddress:  ip,
	})

	return nil
}

// --- MENU ITEMS ---

func (s *MenuService) ListMenuItems(ctx context.Context, search, categoryID string, availableOnly bool, page, perPage int) ([]*models.MenuItem, *models.PaginationMeta, error) {
	items, total, err := s.menuRepo.ListMenuItems(ctx, search, categoryID, availableOnly, page, perPage)
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

	return items, meta, nil
}

func (s *MenuService) GetMenuItemByID(ctx context.Context, id string) (*models.MenuItem, error) {
	return s.menuRepo.GetMenuItemByID(ctx, id)
}

func (s *MenuService) CreateMenuItem(ctx context.Context, req *models.CreateMenuItemRequest, actor *models.User, ip string) (*models.MenuItem, error) {
	slug := generateSlug(req.Name)
	m := &models.MenuItem{
		ID:           "menu-" + uuid.NewString()[:8],
		CategoryID:   req.CategoryID,
		Name:         strings.TrimSpace(req.Name),
		Slug:         slug,
		Description:  req.Description,
		BasePrice:    req.BasePrice,
		IsAvailable:  req.IsAvailable,
		IsBestSeller: req.IsBestSeller,
		IsFeatured:   req.IsFeatured,
		ImageURL:     req.ImageURL,
		SortOrder:    req.SortOrder,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.menuRepo.CreateMenuItem(ctx, m, req.Variants, req.ModifierGroupIDs); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "menu.item_create",
		EntityType: "menu_item",
		EntityID:   &m.ID,
		Details:    fmt.Sprintf(`{"name":"%s","base_price":%d}`, m.Name, m.BasePrice),
		IPAddress:  ip,
	})

	return m, nil
}

func (s *MenuService) UpdateMenuItem(ctx context.Context, id string, req *models.UpdateMenuItemRequest, actor *models.User, ip string) (*models.MenuItem, error) {
	existing, err := s.menuRepo.GetMenuItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.CategoryID = req.CategoryID
	existing.Name = strings.TrimSpace(req.Name)
	existing.Slug = generateSlug(req.Name)
	existing.Description = req.Description
	existing.BasePrice = req.BasePrice
	existing.IsAvailable = req.IsAvailable
	existing.IsBestSeller = req.IsBestSeller
	existing.IsFeatured = req.IsFeatured
	if req.ImageURL != "" {
		existing.ImageURL = req.ImageURL
	}
	existing.SortOrder = req.SortOrder
	existing.UpdatedAt = time.Now().UTC()

	if err := s.menuRepo.UpdateMenuItem(ctx, existing); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "menu.item_update",
		EntityType: "menu_item",
		EntityID:   &existing.ID,
		Details:    fmt.Sprintf(`{"name":"%s","base_price":%d}`, existing.Name, existing.BasePrice),
		IPAddress:  ip,
	})

	return existing, nil
}

func (s *MenuService) ToggleAvailability(ctx context.Context, id string, isAvailable bool, actor *models.User, ip string) error {
	if err := s.menuRepo.ToggleAvailability(ctx, id, isAvailable); err != nil {
		return err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "menu.item_toggle_availability",
		EntityType: "menu_item",
		EntityID:   &id,
		Details:    fmt.Sprintf(`{"is_available":%t}`, isAvailable),
		IPAddress:  ip,
	})

	return nil
}

func (s *MenuService) DeleteMenuItem(ctx context.Context, id string, actor *models.User, ip string) error {
	if err := s.menuRepo.SoftDeleteMenuItem(ctx, id); err != nil {
		return err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "menu.item_delete",
		EntityType: "menu_item",
		EntityID:   &id,
		IPAddress:  ip,
	})

	return nil
}

// Helper untuk menghasilkan slug URL
var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}
