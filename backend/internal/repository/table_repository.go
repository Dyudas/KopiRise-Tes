package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"kopirise/backend/internal/database"
	"kopirise/backend/internal/models"
)

var (
	ErrTableNotFound       = errors.New("meja tidak ditemukan")
	ErrTableInactive       = errors.New("meja ini sedang tidak aktif menerima pesanan")
	ErrTableNumberExists   = errors.New("nomor meja sudah terdaftar dalam sistem")
	ErrInvalidQRToken      = errors.New("token QR tidak valid atau telah dicabut")
)

// TableRepository mengelola operasi database untuk area, meja kafe, dan token QR.
type TableRepository struct {
	db *database.DB
}

// NewTableRepository membuat instance TableRepository baru.
func NewTableRepository(db *database.DB) *TableRepository {
	return &TableRepository{db: db}
}

func (r *TableRepository) ListAreas(ctx context.Context) ([]*models.CafeArea, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description, sort_order FROM cafe_areas ORDER BY sort_order ASC, name ASC;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []*models.CafeArea
	for rows.Next() {
		a := &models.CafeArea{}
		var desc sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &desc, &a.SortOrder); err != nil {
			return nil, err
		}
		if desc.Valid {
			a.Description = desc.String
		}
		areas = append(areas, a)
	}
	return areas, nil
}

func (r *TableRepository) ListTables(ctx context.Context, areaID, status string) ([]*models.CafeTable, error) {
	query := `
		SELECT t.id, t.area_id, a.name, t.table_number, t.capacity, t.status, t.is_active,
		       COALESCE(q.token, '') AS active_token, t.created_at, t.updated_at
		FROM cafe_tables t
		JOIN cafe_areas a ON t.area_id = a.id
		LEFT JOIN table_qr_tokens q ON t.id = q.table_id AND q.is_active = 1
		WHERE 1=1
	`
	var args []interface{}
	if areaID != "" {
		query += " AND t.area_id = ?"
		args = append(args, areaID)
	}
	if status != "" {
		query += " AND t.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY a.sort_order ASC, t.table_number ASC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*models.CafeTable
	for rows.Next() {
		t := &models.CafeTable{}
		var activeInt int
		var createdStr, updatedStr string
		if err := rows.Scan(&t.ID, &t.AreaID, &t.AreaName, &t.TableNumber, &t.Capacity, &t.Status, &activeInt,
			&t.ActiveToken, &createdStr, &updatedStr); err != nil {
			return nil, err
		}
		t.IsActive = activeInt == 1
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
		tables = append(tables, t)
	}
	return tables, nil
}

func (r *TableRepository) GetTableByID(ctx context.Context, id string) (*models.CafeTable, error) {
	query := `
		SELECT t.id, t.area_id, a.name, t.table_number, t.capacity, t.status, t.is_active,
		       COALESCE(q.token, '') AS active_token, t.created_at, t.updated_at
		FROM cafe_tables t
		JOIN cafe_areas a ON t.area_id = a.id
		LEFT JOIN table_qr_tokens q ON t.id = q.table_id AND q.is_active = 1
		WHERE t.id = ?;
	`
	t := &models.CafeTable{}
	var activeInt int
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.AreaID, &t.AreaName, &t.TableNumber, &t.Capacity, &t.Status,
		&activeInt, &t.ActiveToken, &createdStr, &updatedStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}
	t.IsActive = activeInt == 1
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return t, nil
}

func (r *TableRepository) CreateTable(ctx context.Context, t *models.CafeTable) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO cafe_tables (id, area_id, table_number, capacity, status, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.AreaID, t.TableNumber, t.Capacity, "available", nowStr, nowStr)
	if err != nil {
		return err
	}
	return nil
}

func (r *TableRepository) UpdateTable(ctx context.Context, t *models.CafeTable) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	activeInt := 0
	if t.IsActive {
		activeInt = 1
	}
	query := `
		UPDATE cafe_tables
		SET area_id = ?, table_number = ?, capacity = ?, status = ?, is_active = ?, updated_at = ?
		WHERE id = ?;
	`
	res, err := r.db.ExecContext(ctx, query, t.AreaID, t.TableNumber, t.Capacity, t.Status, activeInt, nowStr, t.ID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrTableNotFound
	}
	return nil
}

func (r *TableRepository) UpdateTableStatus(ctx context.Context, tableID, status string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, "UPDATE cafe_tables SET status = ?, updated_at = ? WHERE id = ?;", status, nowStr, tableID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrTableNotFound
	}
	return nil
}

// QR Token Management

func (r *TableRepository) GenerateQRToken(ctx context.Context, tableID, tokenID, tokenStr string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// Cabut token aktif sebelumnya
	_, err = tx.ExecContext(ctx, "UPDATE table_qr_tokens SET is_active = 0, revoked_at = ? WHERE table_id = ? AND is_active = 1;", nowStr, tableID)
	if err != nil {
		return err
	}

	// Buat token baru
	insertSQL := `
		INSERT INTO table_qr_tokens (id, table_id, token, is_active, created_at)
		VALUES (?, ?, ?, 1, ?);
	`
	if _, err := tx.ExecContext(ctx, insertSQL, tokenID, tableID, tokenStr, nowStr); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TableRepository) RevokeQRToken(ctx context.Context, tableID string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, "UPDATE table_qr_tokens SET is_active = 0, revoked_at = ? WHERE table_id = ? AND is_active = 1;", nowStr, tableID)
	return err
}

// GetTableByQRToken menyelesaikan token acak dari scan kamera HP pelanggan ke identitas meja asli.
func (r *TableRepository) GetTableByQRToken(ctx context.Context, token string) (*models.CafeTable, error) {
	query := `
		SELECT t.id, t.area_id, a.name, t.table_number, t.capacity, t.status, t.is_active,
		       q.token, t.created_at, t.updated_at
		FROM table_qr_tokens q
		JOIN cafe_tables t ON q.table_id = t.id
		JOIN cafe_areas a ON t.area_id = a.id
		WHERE q.token = ? AND q.is_active = 1;
	`
	t := &models.CafeTable{}
	var activeInt int
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, token).Scan(&t.ID, &t.AreaID, &t.AreaName, &t.TableNumber, &t.Capacity, &t.Status,
		&activeInt, &t.ActiveToken, &createdStr, &updatedStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidQRToken
		}
		return nil, err
	}

	if activeInt != 1 {
		return nil, ErrTableInactive
	}
	t.IsActive = true
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return t, nil
}
