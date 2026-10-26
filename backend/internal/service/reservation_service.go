package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/repository"
)

var (
	ErrPastDate           = errors.New("tanggal reservasi tidak boleh di masa lampau")
	ErrOutsideHours       = errors.New("waktu reservasi harus berada di dalam jam operasional kafe (08:00 - 22:00 WIB)")
	ErrSpamDetected       = errors.New("permintaan reservasi ditolak oleh filter anti-spam")
	ErrNoAvailableTable   = errors.New("tidak ada meja yang tersedia pada area atau kapasitas tersebut untuk waktu yang dipilih")
)

// ReservationService mengelola logika bisnis reservasi meja dan pencegahan double-booking.
type ReservationService struct {
	reservationRepo *repository.ReservationRepository
	tableRepo       *repository.TableRepository
	audit           *repository.AuditRepository
}

// NewReservationService membuat instance ReservationService baru.
func NewReservationService(resRepo *repository.ReservationRepository, tableRepo *repository.TableRepository, audit *repository.AuditRepository) *ReservationService {
	return &ReservationService{
		reservationRepo: resRepo,
		tableRepo:       tableRepo,
		audit:           audit,
	}
}

// CreatePublicReservation memproses reservasi dari pelanggan dengan proteksi double-booking.
func (s *ReservationService) CreatePublicReservation(ctx context.Context, req *models.CreateReservationRequest, ip string) (*models.Reservation, error) {
	// 1. Cek Honeypot anti-spam
	if strings.TrimSpace(req.WebsiteTrap) != "" {
		return nil, ErrSpamDetected
	}

	// 2. Validasi Tanggal (WIB / Asia/Jakarta)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	nowLocal := time.Now().In(loc)
	todayStr := nowLocal.Format("2006-01-02")

	if req.ReservationDate < todayStr {
		return nil, ErrPastDate
	}

	// 3. Validasi Jam Operasional (08:00 - 22:00 untuk kedatangan terakhir)
	cleanTime := strings.TrimSpace(req.ReservationTime)
	if len(cleanTime) < 5 {
		return nil, errors.New("format jam tidak valid, gunakan format JJ:MM")
	}
	hourMinute := cleanTime[:5]
	if hourMinute < "08:00" || hourMinute > "22:00" {
		return nil, ErrOutsideHours
	}

	durationMinutes := 120 // Standar 2 jam

	var assignedTableID *string
	// 4. Jika Meja spesifik dipilih, lakukan pengecekan bentrok langsung
	if req.TableID != nil && *req.TableID != "" {
		available, err := s.reservationRepo.CheckTableAvailability(ctx, *req.TableID, req.ReservationDate, hourMinute, durationMinutes, "")
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, repository.ErrDoubleBooking
		}
		assignedTableID = req.TableID
	} else {
		// Jika meja belum ditentukan tapi ada areaID atau preferensi kapasitas, cari meja yang cocok dan kosong
		var areaIDFilter string
		if req.AreaID != nil {
			areaIDFilter = *req.AreaID
		}
		tables, err := s.tableRepo.ListTables(ctx, areaIDFilter, "")
		if err != nil {
			return nil, err
		}

		for _, tbl := range tables {
			if tbl.IsActive && tbl.Capacity >= req.GuestCount {
				available, err := s.reservationRepo.CheckTableAvailability(ctx, tbl.ID, req.ReservationDate, hourMinute, durationMinutes, "")
				if err == nil && available {
					assignedTableID = &tbl.ID
					break
				}
			}
		}

		// Jika tidak ada meja yang bisa di-assign secara otomatis, tetap terima sebagai pending tanpa meja (akan dialokasikan manual oleh manager)
	}

	// 5. Generate Nomor Reservasi Unik: RSV-YYYYMMDD-XXXX
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randNum := n.Int64() + 1000
	datePart := strings.ReplaceAll(req.ReservationDate, "-", "")
	rsvNumber := fmt.Sprintf("RSV-%s-%04d", datePart, randNum)

	rsv := &models.Reservation{
		ID:                "rsv-" + uuid.NewString()[:8],
		ReservationNumber: rsvNumber,
		CustomerName:      strings.TrimSpace(req.CustomerName),
		CustomerPhone:     strings.TrimSpace(req.CustomerPhone),
		CustomerEmail:     strings.TrimSpace(req.CustomerEmail),
		GuestCount:        req.GuestCount,
		ReservationDate:   req.ReservationDate,
		ReservationTime:   hourMinute + ":00",
		DurationMinutes:   durationMinutes,
		AreaID:            req.AreaID,
		TableID:           assignedTableID,
		Status:            "pending",
		Notes:             strings.TrimSpace(req.Notes),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	if err := s.reservationRepo.CreateReservation(ctx, rsv); err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		Action:     "reservation.public_create",
		EntityType: "reservation",
		EntityID:   &rsv.ID,
		Details:    fmt.Sprintf(`{"number":"%s","name":"%s","date":"%s"}`, rsv.ReservationNumber, rsv.CustomerName, rsv.ReservationDate),
		IPAddress:  ip,
	})

	return rsv, nil
}

func (s *ReservationService) UpdateStatus(ctx context.Context, id, toStatus, notes string, actor *models.User, ip string) (*models.Reservation, error) {
	var actorID *string
	if actor != nil {
		actorID = &actor.ID
	}

	if err := s.reservationRepo.UpdateReservationStatus(ctx, id, toStatus, notes, actorID); err != nil {
		return nil, err
	}

	updated, err := s.reservationRepo.GetReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.audit.Record(ctx, &models.ActivityLog{
		ID:         uuid.NewString(),
		UserID:     actorID,
		Action:     "reservation.status_update",
		EntityType: "reservation",
		EntityID:   &id,
		Details:    fmt.Sprintf(`{"to_status":"%s","notes":"%s"}`, toStatus, notes),
		IPAddress:  ip,
	})

	return updated, nil
}

func (s *ReservationService) GetReservationByNumber(ctx context.Context, number string) (*models.Reservation, error) {
	return s.reservationRepo.GetReservationByNumber(ctx, number)
}

func (s *ReservationService) ListReservations(ctx context.Context, dateStr, status string, page, perPage int) ([]*models.Reservation, *models.PaginationMeta, error) {
	list, total, err := s.reservationRepo.ListReservations(ctx, dateStr, status, page, perPage)
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

	return list, meta, nil
}
