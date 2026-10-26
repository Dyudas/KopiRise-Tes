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
	ErrUserNotFound    = errors.New("pengguna tidak ditemukan")
	ErrSessionNotFound = errors.New("sesi tidak ditemukan atau telah kedaluwarsa")
	ErrEmailExists     = errors.New("email sudah terdaftar dalam sistem")
)

// UserRepository mengelola operasi database untuk akun pengguna, peran, dan sesi.
type UserRepository struct {
	db *database.DB
}

// NewUserRepository membuat instance UserRepository baru.
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail mengambil data pengguna beserta perannya berdasarkan alamat email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, status, created_at, updated_at, deleted_at
		FROM users
		WHERE email = ? AND deleted_at IS NULL;
	`

	user := &models.User{}
	var phone sql.NullString
	var deletedAt sql.NullString
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&phone,
		&user.PasswordHash,
		&user.Status,
		&createdAtStr,
		&updatedAtStr,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query get by email gagal: %w", err)
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	// Ambil roles
	roles, err := r.getUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	// Ambil profil jika ada
	profile, err := r.getEmployeeProfile(ctx, user.ID)
	if err == nil {
		user.Profile = profile
	}

	return user, nil
}

// GetByID mengambil pengguna berdasarkan ID unik.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, status, created_at, updated_at, deleted_at
		FROM users
		WHERE id = ? AND deleted_at IS NULL;
	`

	user := &models.User{}
	var phone sql.NullString
	var deletedAt sql.NullString
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&phone,
		&user.PasswordHash,
		&user.Status,
		&createdAtStr,
		&updatedAtStr,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query get by id gagal: %w", err)
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	roles, err := r.getUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	profile, err := r.getEmployeeProfile(ctx, user.ID)
	if err == nil {
		user.Profile = profile
	}

	return user, nil
}

// CreateUserWithProfile menyimpan pengguna baru, roles, dan employee profile dalam satu transaksi SQL atomik.
func (r *UserRepository) CreateUserWithProfile(ctx context.Context, user *models.User, profile *models.EmployeeProfile, roles []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// 1. Insert user
	insertUserSQL := `
		INSERT INTO users (id, name, email, phone, password_hash, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, insertUserSQL,
		user.ID,
		user.Name,
		strings.ToLower(strings.TrimSpace(user.Email)),
		user.Phone,
		user.PasswordHash,
		user.Status,
		nowStr,
		nowStr,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return ErrEmailExists
		}
		return fmt.Errorf("insert user gagal: %w", err)
	}

	// 2. Insert roles
	insertRoleSQL := `INSERT INTO user_roles (user_id, role_id, assigned_at) VALUES (?, ?, ?);`
	for _, role := range roles {
		if _, err := tx.ExecContext(ctx, insertRoleSQL, user.ID, role, nowStr); err != nil {
			return fmt.Errorf("insert role %s gagal: %w", role, err)
		}
	}

	// 3. Insert employee profile jika disediakan
	if profile != nil {
		insertProfileSQL := `
			INSERT INTO employee_profiles (user_id, employee_code, id_card_number, date_of_birth, joined_date, hourly_rate, notes, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
		`
		_, err = tx.ExecContext(ctx, insertProfileSQL,
			user.ID,
			profile.EmployeeCode,
			profile.IDCardNumber,
			profile.DateOfBirth,
			profile.JoinedDate,
			profile.HourlyRate,
			profile.Notes,
			nowStr,
			nowStr,
		)
		if err != nil {
			return fmt.Errorf("insert employee profile gagal: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaksi pembuatan user gagal: %w", err)
	}

	return nil
}

// UpdateUser memperbarui info pengguna, peran, dan profil dalam satu transaksi atomik.
func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User, profile *models.EmployeeProfile, roles []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)

	// Update user
	updateUserSQL := `UPDATE users SET name = ?, phone = ?, status = ?, updated_at = ? WHERE id = ?;`
	if _, err := tx.ExecContext(ctx, updateUserSQL, user.Name, user.Phone, user.Status, nowStr, user.ID); err != nil {
		return fmt.Errorf("update user gagal: %w", err)
	}

	// Update roles jika disediakan
	if len(roles) > 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = ?;`, user.ID); err != nil {
			return fmt.Errorf("hapus role lama gagal: %w", err)
		}
		for _, role := range roles {
			if _, err := tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id, assigned_at) VALUES (?, ?, ?);`, user.ID, role, nowStr); err != nil {
				return fmt.Errorf("insert role baru gagal: %w", err)
			}
		}
	}

	// Update profile jika ada
	if profile != nil {
		upsertProfileSQL := `
			INSERT INTO employee_profiles (user_id, employee_code, id_card_number, date_of_birth, joined_date, hourly_rate, notes, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(user_id) DO UPDATE SET
				hourly_rate = excluded.hourly_rate,
				notes = excluded.notes,
				updated_at = excluded.updated_at;
		`
		_, err := tx.ExecContext(ctx, upsertProfileSQL,
			user.ID,
			profile.EmployeeCode,
			profile.IDCardNumber,
			profile.DateOfBirth,
			profile.JoinedDate,
			profile.HourlyRate,
			profile.Notes,
			nowStr,
			nowStr,
		)
		if err != nil {
			return fmt.Errorf("upsert employee profile gagal: %w", err)
		}
	}

	return tx.Commit()
}

// UpdatePassword memperbarui hash kata sandi pengguna.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?;`
	res, err := r.db.ExecContext(ctx, query, passwordHash, nowStr, userID)
	if err != nil {
		return fmt.Errorf("update password gagal: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ListUsers menampilkan daftar pengguna dengan filter dan paginasi server-side.
func (r *UserRepository) ListUsers(ctx context.Context, search, roleFilter, statusFilter string, page, perPage int) ([]*models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "u.deleted_at IS NULL")

	if search != "" {
		whereClauses = append(whereClauses, "(u.name LIKE ? OR u.email LIKE ? OR u.phone LIKE ?)")
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if statusFilter != "" {
		whereClauses = append(whereClauses, "u.status = ?")
		args = append(args, statusFilter)
	}

	if roleFilter != "" {
		whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id AND ur.role_id = ?)")
		args = append(args, roleFilter)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Hitung total items
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM users u WHERE %s;", whereSQL)
	var totalItems int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&totalItems); err != nil {
		return nil, 0, fmt.Errorf("hitung total user gagal: %w", err)
	}

	// Query data
	querySQL := fmt.Sprintf(`
		SELECT u.id, u.name, u.email, u.phone, u.status, u.created_at, u.updated_at
		FROM users u
		WHERE %s
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?;
	`, whereSQL)

	queryArgs := append(args, perPage, offset)
	rows, err := r.db.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query list user gagal: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		var phone sql.NullString
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &phone, &u.Status, &createdAtStr, &updatedAtStr); err != nil {
			return nil, 0, fmt.Errorf("scan user row gagal: %w", err)
		}
		if phone.Valid {
			u.Phone = phone.String
		}
		u.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		// Isi roles
		roles, _ := r.getUserRoles(ctx, u.ID)
		u.Roles = roles

		users = append(users, u)
	}

	return users, totalItems, nil
}

// Session Management

// CreateSession menyimpan sesi baru di database.
func (r *UserRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, ip_address, user_agent, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt.UTC().Format(time.RFC3339),
		session.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("simpan session gagal: %w", err)
	}
	return nil
}

// GetSessionByID mengambil data sesi yang masih berlaku.
func (r *UserRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	query := `
		SELECT id, user_id, ip_address, user_agent, expires_at, created_at
		FROM sessions
		WHERE id = ?;
	`
	s := &models.Session{}
	var ip, ua sql.NullString
	var expiresAtStr, createdAtStr string

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&s.ID,
		&s.UserID,
		&ip,
		&ua,
		&expiresAtStr,
		&createdAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("query session gagal: %w", err)
	}

	if ip.Valid {
		s.IPAddress = ip.String
	}
	if ua.Valid {
		s.UserAgent = ua.String
	}
	s.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

	// Periksa kedaluwarsa
	if time.Now().UTC().After(s.ExpiresAt) {
		_ = r.DeleteSession(ctx, sessionID)
		return nil, ErrSessionNotFound
	}

	return s, nil
}

// DeleteSession mencabut/menghapus sesi tertentu (logout).
func (r *UserRepository) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?;", sessionID)
	return err
}

// Helpers

func (r *UserRepository) getUserRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT role_id FROM user_roles WHERE user_id = ?;", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err == nil {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (r *UserRepository) getEmployeeProfile(ctx context.Context, userID string) (*models.EmployeeProfile, error) {
	query := `
		SELECT user_id, employee_code, id_card_number, date_of_birth, joined_date, hourly_rate, notes, created_at, updated_at
		FROM employee_profiles
		WHERE user_id = ?;
	`
	p := &models.EmployeeProfile{}
	var idCard, dob, notes sql.NullString
	var createdStr, updatedStr string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&p.UserID,
		&p.EmployeeCode,
		&idCard,
		&dob,
		&p.JoinedDate,
		&p.HourlyRate,
		&notes,
		&createdStr,
		&updatedStr,
	)
	if err != nil {
		return nil, err
	}

	if idCard.Valid {
		p.IDCardNumber = idCard.String
	}
	if dob.Valid {
		p.DateOfBirth = dob.String
	}
	if notes.Valid {
		p.Notes = notes.String
	}
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	return p, nil
}
