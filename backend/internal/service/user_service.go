package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/config"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/pkg/argon2id"
	"kopirise/backend/internal/repository"
)

var (
	ErrUnauthorizedAction = errors.New("anda tidak memiliki wewenang untuk tindakan ini")
	ErrInvalidRole        = errors.New("peran pengguna tidak valid")
)

// UserService mengelola operasi pengguna dan data kepegawaian.
type UserService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
	audit    *repository.AuditRepository
}

// NewUserService membuat instance UserService baru.
func NewUserService(cfg *config.Config, userRepo *repository.UserRepository, audit *repository.AuditRepository) *UserService {
	return &UserService{
		cfg:      cfg,
		userRepo: userRepo,
		audit:    audit,
	}
}

// CreateUser membuat akun pengguna baru beserta profil karyawan dan penugasan perannya.
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest, actor *models.User, ipAddress string) (*models.User, error) {
	// Hanya Admin/Owner yang boleh membuat akun
	if !actor.HasRole(models.RoleAdmin) {
		return nil, ErrUnauthorizedAction
	}

	// Validasi peran
	validRoles := map[string]bool{
		models.RoleAdmin:    true,
		models.RoleManager:  true,
		models.RoleCashier:  true,
		models.RoleBarista:  true,
		models.RoleWaiter:   true,
		models.RoleCustomer: true,
	}
	for _, r := range req.Roles {
		if !validRoles[r] {
			return nil, fmt.Errorf("%w: %s", ErrInvalidRole, r)
		}
	}

	// Hash password menggunakan Argon2id
	argonParams := &argon2id.Params{
		Memory:      s.cfg.Argon2Memory,
		Iterations:  s.cfg.Argon2Iterations,
		Parallelism: s.cfg.Argon2Parallelism,
		SaltLength:  s.cfg.Argon2SaltLength,
		KeyLength:   s.cfg.Argon2KeyLength,
	}
	passwordHash, err := argon2id.HashPassword(req.Password, argonParams)
	if err != nil {
		return nil, fmt.Errorf("hashing password gagal: %w", err)
	}

	userID := "usr-" + uuid.NewString()[:8]
	newUser := &models.User{
		ID:           userID,
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:        strings.TrimSpace(req.Phone),
		PasswordHash: passwordHash,
		Status:       "active",
		Roles:        req.Roles,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	var profile *models.EmployeeProfile
	if req.EmployeeCode != "" || req.HourlyRate > 0 || req.JoinedDate != "" {
		empCode := req.EmployeeCode
		if empCode == "" {
			empCode = "EMP-" + strings.ToUpper(uuid.NewString()[:6])
		}
		joined := req.JoinedDate
		if joined == "" {
			joined = time.Now().UTC().Format("2006-01-02")
		}

		profile = &models.EmployeeProfile{
			UserID:       userID,
			EmployeeCode: empCode,
			JoinedDate:   joined,
			HourlyRate:   req.HourlyRate,
			Notes:        req.Notes,
		}
		newUser.Profile = profile
	}

	if err := s.userRepo.CreateUserWithProfile(ctx, newUser, profile, req.Roles); err != nil {
		return nil, err
	}

	// Audit log
	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "user.create",
		EntityType: "user",
		EntityID:   &newUser.ID,
		Details:    fmt.Sprintf(`{"email":"%s","roles":%q}`, newUser.Email, newUser.Roles),
		IPAddress:  ipAddress,
		CreatedAt:  time.Now().UTC(),
	})

	return newUser, nil
}

// UpdateUser memperbarui data profil pengguna dan status.
func (s *UserService) UpdateUser(ctx context.Context, targetID string, req *models.UpdateUserRequest, actor *models.User, ipAddress string) (*models.User, error) {
	target, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	// Proteksi: Manager tidak boleh mengubah data Admin/Owner
	if target.HasRole(models.RoleAdmin) && !actor.HasRole(models.RoleAdmin) {
		return nil, ErrUnauthorizedAction
	}

	// Proteksi: Hanya admin yang boleh mengubah peran akun
	if len(req.Roles) > 0 && !actor.HasRole(models.RoleAdmin) {
		return nil, ErrUnauthorizedAction
	}

	target.Name = strings.TrimSpace(req.Name)
	target.Phone = strings.TrimSpace(req.Phone)
	target.Status = req.Status
	target.UpdatedAt = time.Now().UTC()

	var profile *models.EmployeeProfile
	if target.Profile != nil {
		profile = target.Profile
		if req.HourlyRate != nil {
			profile.HourlyRate = *req.HourlyRate
		}
		if req.Notes != nil {
			profile.Notes = *req.Notes
		}
	}

	if err := s.userRepo.UpdateUser(ctx, target, profile, req.Roles); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "user.update",
		EntityType: "user",
		EntityID:   &target.ID,
		Details:    fmt.Sprintf(`{"status":"%s"}`, target.Status),
		IPAddress:  ipAddress,
		CreatedAt:  time.Now().UTC(),
	})

	return target, nil
}

// ResetPassword mereset password akun pengguna dengan hash baru.
func (s *UserService) ResetPassword(ctx context.Context, targetID string, newPassword string, actor *models.User, ipAddress string) error {
	target, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	// Manager tidak boleh mereset password Admin/Owner
	if target.HasRole(models.RoleAdmin) && !actor.HasRole(models.RoleAdmin) {
		return ErrUnauthorizedAction
	}

	argonParams := &argon2id.Params{
		Memory:      s.cfg.Argon2Memory,
		Iterations:  s.cfg.Argon2Iterations,
		Parallelism: s.cfg.Argon2Parallelism,
		SaltLength:  s.cfg.Argon2SaltLength,
		KeyLength:   s.cfg.Argon2KeyLength,
	}
	hash, err := argon2id.HashPassword(newPassword, argonParams)
	if err != nil {
		return fmt.Errorf("hashing password baru gagal: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, targetID, hash); err != nil {
		return err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     &actor.ID,
		Action:     "user.reset_password",
		EntityType: "user",
		EntityID:   &targetID,
		IPAddress:  ipAddress,
		CreatedAt:  time.Now().UTC(),
	})

	return nil
}

// ListUsers menampilkan daftar pengguna dengan paginasi metadata.
func (s *UserService) ListUsers(ctx context.Context, search, roleFilter, statusFilter string, page, perPage int) ([]*models.User, *models.PaginationMeta, error) {
	users, totalItems, err := s.userRepo.ListUsers(ctx, search, roleFilter, statusFilter, page, perPage)
	if err != nil {
		return nil, nil, err
	}

	totalPages := int((totalItems + int64(perPage) - 1) / int64(perPage))
	meta := &models.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return users, meta, nil
}
