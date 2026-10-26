package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/config"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/pkg/argon2id"
	"kopirise/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("email atau kata sandi tidak valid")
	ErrAccountInactive    = errors.New("akun Anda sedang dinonaktifkan atau ditangguhkan")
)

// AuthService mengelola logika bisnis autentikasi dan manajemen sesi.
type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
	audit    *repository.AuditRepository
}

// NewAuthService membuat instance AuthService.
func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository, audit *repository.AuditRepository) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		audit:    audit,
	}
}

// Login memverifikasi kredensial pengguna dan menerbitkan sesi baru.
func (s *AuthService) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*models.User, *models.Session, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		slog.Warn("Upaya login gagal: email tidak ditemukan", "email", email, "ip", ipAddress)
		return nil, nil, ErrInvalidCredentials
	}

	if user.Status != "active" {
		slog.Warn("Upaya login akun nonaktif", "userId", user.ID, "status", user.Status, "ip", ipAddress)
		return nil, nil, ErrAccountInactive
	}

	match, err := argon2id.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		slog.Warn("Upaya login gagal: password salah", "userId", user.ID, "ip", ipAddress)
		return nil, nil, ErrInvalidCredentials
	}

	// Terbitkan token sesi kriptografis acak 32 bytes (64 karakter hex)
	sessionTokenBytes := make([]byte, 32)
	if _, err := rand.Read(sessionTokenBytes); err != nil {
		return nil, nil, fmt.Errorf("gagal menghasilkan token sesi: %w", err)
	}
	sessionID := hex.EncodeToString(sessionTokenBytes)

	now := time.Now().UTC()
	session := &models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: now.Add(s.cfg.SessionExpiry),
		CreatedAt: now,
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("gagal menyimpan sesi: %w", err)
	}

	// Catat audit log login
	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &user.ID,
		Action:     "auth.login",
		EntityType: "session",
		EntityID:   &session.ID,
		Details:    fmt.Sprintf(`{"email":"%s","roles":%q}`, user.Email, user.Roles),
		IPAddress:  ipAddress,
		CreatedAt:  now,
	})

	slog.Info("Login berhasil", "userId", user.ID, "roles", user.Roles, "ip", ipAddress)
	return user, session, nil
}

// Logout mencabut sesi yang aktif di database.
func (s *AuthService) Logout(ctx context.Context, sessionID string, userID *string, ipAddress string) error {
	if sessionID == "" {
		return nil
	}

	if err := s.userRepo.DeleteSession(ctx, sessionID); err != nil {
		slog.Error("Gagal menghapus sesi saat logout", "sessionId", sessionID, "error", err)
		return err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     userID,
		Action:     "auth.logout",
		EntityType: "session",
		EntityID:   &sessionID,
		IPAddress:  ipAddress,
		CreatedAt:  time.Now().UTC(),
	})

	return nil
}

// GetUserFromSession mengambil pengguna yang terafiliasi dengan sesi aktif.
func (s *AuthService) GetUserFromSession(ctx context.Context, sessionID string) (*models.User, error) {
	session, err := s.userRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrAccountInactive
	}

	return user, nil
}
