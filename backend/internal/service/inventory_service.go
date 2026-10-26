package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
)

type InventoryService struct {
	inventoryRepo *repository.InventoryRepository
	auditRepo     *repository.AuditRepository
}

func NewInventoryService(inventoryRepo *repository.InventoryRepository, auditRepo *repository.AuditRepository) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		auditRepo:     auditRepo,
	}
}

// 1. Bahan Baku (Ingredients)
func (s *InventoryService) ListIngredients(ctx context.Context) ([]*models.Ingredient, error) {
	return s.inventoryRepo.ListIngredients(ctx)
}

func (s *InventoryService) ListLowStock(ctx context.Context) ([]*models.Ingredient, error) {
	return s.inventoryRepo.ListLowStockIngredients(ctx)
}

func (s *InventoryService) GetIngredientByID(ctx context.Context, id string) (*models.Ingredient, error) {
	return s.inventoryRepo.GetIngredientByID(ctx, id)
}

func (s *InventoryService) CreateIngredient(ctx context.Context, req *models.CreateIngredientRequest, actor *models.User, ip string) (*models.Ingredient, error) {
	ing := &models.Ingredient{
		ID:           "ing-" + uuid.New().String()[:8],
		Name:         req.Name,
		Unit:         req.Unit,
		CurrentStock: 0,
		MinimumStock: req.MinimumStock,
		CostPerUnit:  req.CostPerUnit,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.inventoryRepo.CreateIngredient(ctx, ing); err != nil {
		return nil, err
	}

	_ = s.auditRepo.Log(ctx, &models.ActivityLog{
		UserID:       &actor.ID,
		Action:       "CREATE_INGREDIENT",
		ResourceType: "ingredients",
		ResourceID:   &ing.ID,
		Details:      fmt.Sprintf("Menambah bahan baku: %s (%s)", ing.Name, ing.Unit),
		IPAddress:    ip,
	})

	return ing, nil
}

func (s *InventoryService) UpdateIngredient(ctx context.Context, id string, req *models.CreateIngredientRequest, actor *models.User, ip string) error {
	ing := &models.Ingredient{
		ID:           id,
		Name:         req.Name,
		Unit:         req.Unit,
		MinimumStock: req.MinimumStock,
		IsActive:     true,
	}

	if err := s.inventoryRepo.UpdateIngredient(ctx, ing); err != nil {
		return err
	}

	_ = s.auditRepo.Log(ctx, &models.ActivityLog{
		UserID:       &actor.ID,
		Action:       "UPDATE_INGREDIENT",
		ResourceType: "ingredients",
		ResourceID:   &id,
		Details:      fmt.Sprintf("Mengubah data bahan baku ID: %s", id),
		IPAddress:    ip,
	})

	return nil
}

// 2. Supplier
func (s *InventoryService) ListSuppliers(ctx context.Context) ([]*models.Supplier, error) {
	return s.inventoryRepo.ListSuppliers(ctx)
}

func (s *InventoryService) CreateSupplier(ctx context.Context, req *models.CreateSupplierRequest, actor *models.User, ip string) (*models.Supplier, error) {
	sup := &models.Supplier{
		ID:            "sup-" + uuid.New().String()[:8],
		Name:          req.Name,
		Phone:         req.Phone,
		Email:         req.Email,
		Address:       req.Address,
		ContactPerson: req.ContactPerson,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.inventoryRepo.CreateSupplier(ctx, sup); err != nil {
		return nil, err
	}

	_ = s.auditRepo.Log(ctx, &models.ActivityLog{
		UserID:       &actor.ID,
		Action:       "CREATE_SUPPLIER",
		ResourceType: "suppliers",
		ResourceID:   &sup.ID,
		Details:      fmt.Sprintf("Menambah data supplier: %s", sup.Name),
		IPAddress:    ip,
	})

	return sup, nil
}

// 3. Pembelian & Pengadaan Bahan Baku (Purchases)
func (s *InventoryService) CreatePurchase(ctx context.Context, req *models.CreatePurchaseRequest, actor *models.User, ip string) (*models.IngredientPurchase, error) {
	var totalAmount int64
	var purchaseItems []*models.IngredientPurchaseItem

	purchaseID := "po-" + uuid.New().String()[:8]
	purchaseNumber := generatePurchaseNumber()

	for _, itemReq := range req.Items {
		subtotal := int64(itemReq.Quantity * float64(itemReq.UnitCost))
		totalAmount += subtotal

		purchaseItems = append(purchaseItems, &models.IngredientPurchaseItem{
			ID:           "poi-" + uuid.New().String()[:8],
			PurchaseID:   purchaseID,
			IngredientID: itemReq.IngredientID,
			Quantity:     itemReq.Quantity,
			UnitCost:     itemReq.UnitCost,
			Subtotal:     subtotal,
		})
	}

	purchase := &models.IngredientPurchase{
		ID:                 purchaseID,
		PurchaseNumber:     purchaseNumber,
		SupplierID:         req.SupplierID,
		FinancialAccountID: req.FinancialAccountID,
		PurchaseDate:       req.PurchaseDate,
		TotalAmount:        totalAmount,
		CreatedByUserID:    actor.ID,
		Notes:              req.Notes,
		Items:              purchaseItems,
		CreatedAt:          time.Now().UTC(),
	}

	if err := s.inventoryRepo.CreatePurchase(ctx, purchase, purchaseItems); err != nil {
		return nil, err
	}

	_ = s.auditRepo.Log(ctx, &models.ActivityLog{
		UserID:       &actor.ID,
		Action:       "CREATE_PURCHASE",
		ResourceType: "ingredient_purchases",
		ResourceID:   &purchase.ID,
		Details:      fmt.Sprintf("Pembelian bahan baku #%s total Rp %d", purchase.PurchaseNumber, purchase.TotalAmount),
		IPAddress:    ip,
	})

	return purchase, nil
}

func (s *InventoryService) ListPurchases(ctx context.Context, page, perPage int) ([]*models.IngredientPurchase, *models.PaginationMeta, error) {
	return s.inventoryRepo.ListPurchases(ctx, page, perPage)
}

// 4. Resep Menu (Bill of Materials) & HPP
func (s *InventoryService) GetRecipe(ctx context.Context, menuItemID string) (*models.Recipe, error) {
	return s.inventoryRepo.GetRecipeByMenuItemID(ctx, menuItemID)
}

func (s *InventoryService) SaveRecipe(ctx context.Context, req *models.CreateRecipeRequest, actor *models.User, ip string) (*models.Recipe, error) {
	recipeID := "rec-" + uuid.New().String()[:8]
	var items []*models.RecipeItem

	for _, itm := range req.Items {
		items = append(items, &models.RecipeItem{
			ID:           "ri-" + uuid.New().String()[:8],
			RecipeID:     recipeID,
			IngredientID: itm.IngredientID,
			Quantity:     itm.Quantity,
		})
	}

	recipe := &models.Recipe{
		ID:         recipeID,
		MenuItemID: req.MenuItemID,
		Notes:      req.Notes,
		Items:      items,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.inventoryRepo.CreateOrUpdateRecipe(ctx, recipe, items); err != nil {
		return nil, err
	}

	_ = s.auditRepo.Log(ctx, &models.ActivityLog{
		UserID:       &actor.ID,
		Action:       "SAVE_RECIPE",
		ResourceType: "recipes",
		ResourceID:   &recipe.ID,
		Details:      fmt.Sprintf("Menyimpan resep BOM menu ID: %s", req.MenuItemID),
		IPAddress:    ip,
	})

	return s.inventoryRepo.GetRecipeByMenuItemID(ctx, req.MenuItemID)
}

func (s *InventoryService) DeductStockForCompletedOrder(ctx context.Context, orderID string, actor *models.User) error {
	var actorID *string
	if actor != nil {
		actorID = &actor.ID
	}
	return s.inventoryRepo.DeductStockForOrder(ctx, orderID, actorID)
}

func generatePurchaseNumber() string {
	today := time.Now().Format("20060102")
	return fmt.Sprintf("PO-%s-%s", today, strings.ToUpper(uuid.New().String()[:4]))
}
