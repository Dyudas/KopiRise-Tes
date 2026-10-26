package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
)

// TableService mengelola operasi bisnis meja, area, dan penerbitan token QR Code.
type TableService struct {
	tableRepo *repository.TableRepository
	audit     *repository.AuditRepository
}

// NewTableService membuat instance TableService baru.
func NewTableService(tableRepo *repository.TableRepository, audit *repository.AuditRepository) *TableService {
	return &TableService{
		tableRepo: tableRepo,
		audit:     audit,
	}
}

func (s *TableService) ListAreas(ctx context.Context) ([]*models.CafeArea, error) {
	return s.tableRepo.ListAreas(ctx)
}

func (s *TableService) ListTables(ctx context.Context, areaID, status string) ([]*models.CafeTable, error) {
	return s.tableRepo.ListTables(ctx, areaID, status)
}

func (s *TableService) GetTableByID(ctx context.Context, id string) (*models.CafeTable, error) {
	return s.tableRepo.GetTableByID(ctx, id)
}

func (s *TableService) CreateTable(ctx context.Context, req *models.CreateTableRequest, actor *models.User, ip string) (*models.CafeTable, error) {
	tableID := "tbl-" + strings.ToLower(strings.ReplaceAll(req.TableNumber, " ", "-"))
	t := &models.CafeTable{
		ID:          tableID,
		AreaID:      req.AreaID,
		TableNumber: strings.ToUpper(strings.TrimSpace(req.TableNumber)),
		Capacity:    req.Capacity,
		Status:      "available",
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.tableRepo.CreateTable(ctx, t); err != nil {
		return nil, err
	}

	// Terbitkan QR token awal secara otomatis
	_, _ = s.GenerateQRToken(ctx, t.ID, actor, ip)

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "table.create",
		EntityType: "cafe_table",
		EntityID:   &t.ID,
		Details:    fmt.Sprintf(`{"table_number":"%s","capacity":%d}`, t.TableNumber, t.Capacity),
		IPAddress:  ip,
	})

	return t, nil
}

func (s *TableService) UpdateTable(ctx context.Context, id string, req *models.UpdateTableRequest, actor *models.User, ip string) (*models.CafeTable, error) {
	existing, err := s.tableRepo.GetTableByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.AreaID = req.AreaID
	existing.TableNumber = strings.ToUpper(strings.TrimSpace(req.TableNumber))
	existing.Capacity = req.Capacity
	existing.Status = req.Status
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now().UTC()

	if err := s.tableRepo.UpdateTable(ctx, existing); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "table.update",
		EntityType: "cafe_table",
		EntityID:   &existing.ID,
		Details:    fmt.Sprintf(`{"table_number":"%s","status":"%s"}`, existing.TableNumber, existing.Status),
		IPAddress:  ip,
	})

	return existing, nil
}

func (s *TableService) GenerateQRToken(ctx context.Context, tableID string, actor *models.User, ip string) (string, error) {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("gagal menghasilkan token acak: %w", err)
	}
	tokenStr := "kr_tbl_" + hex.EncodeToString(tokenBytes)
	tokenID := "tok-" + uuid.NewString()[:8]

	if err := s.tableRepo.GenerateQRToken(ctx, tableID, tokenID, tokenStr); err != nil {
		return "", err
	}

	if actor != nil {
		_ = s.audit.Record(ctx, &models.ActivityLog{
			ID:         uuid.NewString(),
			UserID:     &actor.ID,
			Action:     "table.qr_generate",
			EntityType: "table_qr_token",
			EntityID:   &tokenID,
			Details:    fmt.Sprintf(`{"table_id":"%s"}`, tableID),
			IPAddress:  ip,
		})
	}

	return tokenStr, nil
}

func (s *TableService) RevokeQRToken(ctx context.Context, tableID string, actor *models.User, ip string) error {
	if err := s.tableRepo.RevokeQRToken(ctx, tableID); err != nil {
		return err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "table.qr_revoke",
		EntityType: "table_qr_token",
		EntityID:   &tableID,
		IPAddress:  ip,
	})

	return nil
}

// ResolvePublicQR memverifikasi token QR pelanggan dan mengembalikan informasi meja yang valid.
func (s *TableService) ResolvePublicQR(ctx context.Context, token string) (*models.PublicTableQRResponse, error) {
	table, err := s.tableRepo.GetTableByQRToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &models.PublicTableQRResponse{
		TableNumber: table.TableNumber,
		AreaName:    table.AreaName,
		Capacity:    table.Capacity,
		Status:      table.Status,
		Token:       token,
	}, nil
}
