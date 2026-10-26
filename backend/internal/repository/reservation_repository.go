package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/database"
	"kopirise/backend/internal/models"
)

var (
	ErrReservationNotFound = errors.New("reservasi tidak ditemukan")
	ErrDoubleBooking       = errors.New("meja ini sudah dipesan oleh pelanggan lain pada rentang waktu yang dipilih")
	ErrInvalidStatusChange = errors.New("perubahan status reservasi tidak valid")
)

// ReservationRepository mengelola operasi database untuk kalender dan data reservasi meja.
type ReservationRepository struct {
	db *database.DB
}

// NewReservationRepository membuat instance ReservationRepository baru.
func NewReservationRepository(db *database.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// CheckTableAvailability memverifikasi bahwa meja tidak bentrok dengan reservasi terkonfirmasi lain.
func (r *ReservationRepository) CheckTableAvailability(ctx context.Context, tableID, dateStr, timeStr string, durationMinutes int, excludeReservationID string) (bool, error) {
	parsedTime, err := time.Parse("15:04", timeStr[:5])
	if err != nil {
		return false, fmt.Errorf("format jam tidak valid: %w", err)
	}

	newStartMinutes := parsedTime.Hour()*60 + parsedTime.Minute()
	newEndMinutes := newStartMinutes + durationMinutes

	query := `
		SELECT id, reservation_time, duration_minutes
		FROM reservations
		WHERE table_id = ? 
		  AND reservation_date = ? 
		  AND status IN ('confirmed', 'seated')
	`
	var args []interface{}
	args = append(args, tableID, dateStr)

	if excludeReservationID != "" {
		query += " AND id != ?"
		args = append(args, excludeReservationID)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, existTimeStr string
		var existDuration int
		if err := rows.Scan(&id, &existTimeStr, &existDuration); err != nil {
			return false, err
		}

		existTime, err := time.Parse("15:04", existTimeStr[:5])
		if err != nil {
			continue
		}

		existStartMinutes := existTime.Hour()*60 + existTime.Minute()
		existEndMinutes := existStartMinutes + existDuration

		// Cek apakah ada tabrakan interval waktu (overlap check)
		// Tabrakan terjadi jika: newStart < existEnd DAN newEnd > existStart
		if newStartMinutes < existEndMinutes && newEndMinutes > existStartMinutes {
			return false, nil // Meja bentrok / tidak tersedia
		}
	}

	return true, nil
}

// CreateReservation mencatat data reservasi baru dan histori status awal dalam transaksi atomik.
func (r *ReservationRepository) CreateReservation(ctx context.Context, res *models.Reservation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	insertSQL := `
		INSERT INTO reservations (
			id, reservation_number, customer_name, customer_phone, customer_email,
			guest_count, reservation_date, reservation_time, duration_minutes,
			area_id, table_id, status, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertSQL,
		res.ID,
		res.ReservationNumber,
		res.CustomerName,
		res.CustomerPhone,
		res.CustomerEmail,
		res.GuestCount,
		res.ReservationDate,
		res.ReservationTime,
		res.DurationMinutes,
		res.AreaID,
		res.TableID,
		res.Status,
		res.Notes,
		nowStr,
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("insert reservation gagal: %w", err)
	}

	// Catat riwayat status awal
	historySQL := `
		INSERT INTO reservation_status_history (id, reservation_id, from_status, to_status, notes, created_at)
		VALUES (?, ?, NULL, ?, 'Reservasi baru diajukan via website publik', ?);
	`
	historyID := uuid.NewString()
	if _, err := tx.ExecContext(ctx, historySQL, historyID, res.ID, res.Status, nowStr); err != nil {
		return fmt.Errorf("insert reservation status history gagal: %w", err)
	}

	return tx.Commit()
}

// UpdateReservationStatus memperbarui status reservasi, mencatat riwayat, dan menyinkronkan status meja.
func (r *ReservationRepository) UpdateReservationStatus(ctx context.Context, id, toStatus, notes string, userID *string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// Ambil status saat ini dan table_id
	var currentStatus string
	var tableID sql.NullString
	query := `SELECT status, table_id FROM reservations WHERE id = ?;`
	if err := tx.QueryRowContext(ctx, query, id).Scan(&currentStatus, &tableID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrReservationNotFound
		}
		return err
	}

	// Update status reservasi
	updateSQL := `UPDATE reservations SET status = ?, updated_at = ? WHERE id = ?;`
	if _, err := tx.ExecContext(ctx, updateSQL, toStatus, nowStr, id); err != nil {
		return err
	}

	// Catat ke riwayat
	histSQL := `
		INSERT INTO reservation_status_history (id, reservation_id, from_status, to_status, notes, changed_by_user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	histID := uuid.NewString()
	if _, err := tx.ExecContext(ctx, histSQL, histID, id, currentStatus, toStatus, notes, userID, nowStr); err != nil {
		return err
	}

	// Sinkronisasi status fisik meja jika terkait
	if tableID.Valid && tableID.String != "" {
		var newTableStatus string
		switch toStatus {
		case "seated":
			newTableStatus = "occupied"
		case "completed", "cancelled", "no_show":
			newTableStatus = "available"
		}

		if newTableStatus != "" {
			_, _ = tx.ExecContext(ctx, "UPDATE cafe_tables SET status = ?, updated_at = ? WHERE id = ?;", newTableStatus, nowStr, tableID.String)
		}
	}

	return tx.Commit()
}

func (r *ReservationRepository) GetReservationByID(ctx context.Context, id string) (*models.Reservation, error) {
	query := `
		SELECT r.id, r.reservation_number, r.customer_name, r.customer_phone, r.customer_email,
		       r.guest_count, r.reservation_date, r.reservation_time, r.duration_minutes,
		       r.area_id, COALESCE(a.name, '') AS area_name,
		       r.table_id, COALESCE(t.table_number, '') AS table_number,
		       r.status, r.notes, r.created_at, r.updated_at
		FROM reservations r
		LEFT JOIN cafe_areas a ON r.area_id = a.id
		LEFT JOIN cafe_tables t ON r.table_id = t.id
		WHERE r.id = ?;
	`
	res := &models.Reservation{}
	var email, notes, areaID, tableID sql.NullString
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&res.ID, &res.ReservationNumber, &res.CustomerName, &res.CustomerPhone, &email,
		&res.GuestCount, &res.ReservationDate, &res.ReservationTime, &res.DurationMinutes,
		&areaID, &res.AreaName, &tableID, &res.TableNumber, &res.Status, &notes,
		&createdStr, &updatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReservationNotFound
		}
		return nil, err
	}

	if email.Valid {
		res.CustomerEmail = email.String
	}
	if notes.Valid {
		res.Notes = notes.String
	}
	if areaID.Valid {
		res.AreaID = &areaID.String
	}
	if tableID.Valid {
		res.TableID = &tableID.String
	}
	res.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	res.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	// Fetch status history
	res.History, _ = r.getReservationHistory(ctx, res.ID)

	return res, nil
}

func (r *ReservationRepository) GetReservationByNumber(ctx context.Context, number string) (*models.Reservation, error) {
	query := `
		SELECT r.id, r.reservation_number, r.customer_name, r.customer_phone, r.customer_email,
		       r.guest_count, r.reservation_date, r.reservation_time, r.duration_minutes,
		       r.area_id, COALESCE(a.name, '') AS area_name,
		       r.table_id, COALESCE(t.table_number, '') AS table_number,
		       r.status, r.notes, r.created_at, r.updated_at
		FROM reservations r
		LEFT JOIN cafe_areas a ON r.area_id = a.id
		LEFT JOIN cafe_tables t ON r.table_id = t.id
		WHERE r.reservation_number = ?;
	`
	res := &models.Reservation{}
	var email, notes, areaID, tableID sql.NullString
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, number).Scan(
		&res.ID, &res.ReservationNumber, &res.CustomerName, &res.CustomerPhone, &email,
		&res.GuestCount, &res.ReservationDate, &res.ReservationTime, &res.DurationMinutes,
		&areaID, &res.AreaName, &tableID, &res.TableNumber, &res.Status, &notes,
		&createdStr, &updatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReservationNotFound
		}
		return nil, err
	}

	if email.Valid {
		res.CustomerEmail = email.String
	}
	if notes.Valid {
		res.Notes = notes.String
	}
	if areaID.Valid {
		res.AreaID = &areaID.String
	}
	if tableID.Valid {
		res.TableID = &tableID.String
	}
	res.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	res.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	res.History, _ = r.getReservationHistory(ctx, res.ID)

	return res, nil
}

func (r *ReservationRepository) ListReservations(ctx context.Context, dateStr, status string, page, perPage int) ([]*models.Reservation, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	queryWhere := "WHERE 1=1"
	var args []interface{}
	if dateStr != "" {
		queryWhere += " AND r.reservation_date = ?"
		args = append(args, dateStr)
	}
	if status != "" {
		queryWhere += " AND r.status = ?"
		args = append(args, status)
	}

	var totalItems int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM reservations r %s;", queryWhere)
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&totalItems); err != nil {
		return nil, 0, err
	}

	selectSQL := fmt.Sprintf(`
		SELECT r.id, r.reservation_number, r.customer_name, r.customer_phone, r.customer_email,
		       r.guest_count, r.reservation_date, r.reservation_time, r.duration_minutes,
		       r.area_id, COALESCE(a.name, '') AS area_name,
		       r.table_id, COALESCE(t.table_number, '') AS table_number,
		       r.status, r.notes, r.created_at, r.updated_at
		FROM reservations r
		LEFT JOIN cafe_areas a ON r.area_id = a.id
		LEFT JOIN cafe_tables t ON r.table_id = t.id
		%s
		ORDER BY r.reservation_date ASC, r.reservation_time ASC
		LIMIT ? OFFSET ?;
	`, queryWhere)

	queryArgs := append(args, perPage, offset)
	rows, err := r.db.QueryContext(ctx, selectSQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.Reservation
	for rows.Next() {
		res := &models.Reservation{}
		var email, notes, areaID, tableID sql.NullString
		var createdStr, updatedStr string

		if err := rows.Scan(
			&res.ID, &res.ReservationNumber, &res.CustomerName, &res.CustomerPhone, &email,
			&res.GuestCount, &res.ReservationDate, &res.ReservationTime, &res.DurationMinutes,
			&areaID, &res.AreaName, &tableID, &res.TableNumber, &res.Status, &notes,
			&createdStr, &updatedStr,
		); err != nil {
			return nil, 0, err
		}

		if email.Valid {
			res.CustomerEmail = email.String
		}
		if notes.Valid {
			res.Notes = notes.String
		}
		if areaID.Valid {
			res.AreaID = &areaID.String
		}
		if tableID.Valid {
			res.TableID = &tableID.String
		}
		res.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		res.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

		list = append(list, res)
	}

	return list, totalItems, nil
}

func (r *ReservationRepository) getReservationHistory(ctx context.Context, reservationID string) ([]*models.ReservationStatusHistory, error) {
	query := `
		SELECT h.id, h.reservation_id, h.from_status, h.to_status, h.notes,
		       h.changed_by_user_id, COALESCE(u.name, '') AS user_name, h.created_at
		FROM reservation_status_history h
		LEFT JOIN users u ON h.changed_by_user_id = u.id
		WHERE h.reservation_id = ?
		ORDER BY h.created_at ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.ReservationStatusHistory
	for rows.Next() {
		h := &models.ReservationStatusHistory{}
		var fromStatus, notes, userID sql.NullString
		var createdStr string
		if err := rows.Scan(&h.ID, &h.ReservationID, &fromStatus, &h.ToStatus, &notes, &userID, &h.ChangedByUserName, &createdStr); err != nil {
			return nil, err
		}
		if fromStatus.Valid {
			h.FromStatus = fromStatus.String
		}
		if notes.Valid {
			h.Notes = notes.String
		}
		if userID.Valid {
			h.ChangedByUserID = &userID.String
		}
		h.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		history = append(history, h)
	}
	return history, nil
}
