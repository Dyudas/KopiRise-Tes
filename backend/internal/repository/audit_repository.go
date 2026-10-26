package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"kopirise/backend/internal/database"
	"kopirise/backend/internal/models"
)

// AuditRepository mengelola pencatatan dan pembacaan log aktivitas sensitif.
type AuditRepository struct {
	db *database.DB
}

// NewAuditRepository membuat instance AuditRepository baru.
func NewAuditRepository(db *database.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Record mencatat tindakan aktivitas baru ke tabel activity_logs.
func (r *AuditRepository) Record(ctx context.Context, log *models.ActivityLog) error {
	query := `
		INSERT INTO activity_logs (id, user_id, action, entity_type, entity_id, details, ip_address, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`
	nowStr := time.Now().UTC().Format(time.RFC3339)
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	} else {
		nowStr = log.CreatedAt.UTC().Format(time.RFC3339)
	}

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.Details,
		log.IPAddress,
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("gagal mencatat activity log: %w", err)
	}
	return nil
}

// List mengambil riwayat log aktivitas terbaru dengan paginasi.
func (r *AuditRepository) List(ctx context.Context, page, perPage int) ([]*models.ActivityLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var totalItems int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM activity_logs;").Scan(&totalItems); err != nil {
		return nil, 0, fmt.Errorf("hitung total activity log gagal: %w", err)
	}

	query := `
		SELECT id, user_id, action, entity_type, entity_id, details, ip_address, created_at
		FROM activity_logs
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?;
	`
	rows, err := r.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query activity log gagal: %w", err)
	}
	defer rows.Close()

	var logs []*models.ActivityLog
	for rows.Next() {
		l := &models.ActivityLog{}
		var uid, eid, det, ip sql.NullString
		var createdStr string

		if err := rows.Scan(&l.ID, &uid, &l.Action, &l.EntityType, &eid, &det, &ip, &createdStr); err != nil {
			return nil, 0, err
		}
		if uid.Valid {
			l.UserID = &uid.String
		}
		if eid.Valid {
			l.EntityID = &eid.String
		}
		if det.Valid {
			l.Details = det.String
		}
		if ip.Valid {
			l.IPAddress = ip.String
		}
		l.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)

		logs = append(logs, l)
	}

	return logs, totalItems, nil
}
