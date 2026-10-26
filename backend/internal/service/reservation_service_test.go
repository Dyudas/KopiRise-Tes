package service

import (
	"context"
	"testing"
	"time"

	"kopirise/backend/internal/models"
)

func TestHoneypotSpamDetection(t *testing.T) {
	svc := &ReservationService{}

	req := &models.CreateReservationRequest{
		CustomerName:  "Spam Bot",
		CustomerPhone: "081234567890",
		WebsiteTrap:   "http://spamsite.xyz", // Bot fills hidden field
	}

	_, err := svc.CreatePublicReservation(context.Background(), req, "127.0.0.1")
	if err != ErrSpamDetected {
		t.Fatalf("Ekspektasi ErrSpamDetected, didapat: %v", err)
	}
}

func TestOperatingHoursValidation(t *testing.T) {
	svc := &ReservationService{}

	// Reservasi jam 03:00 subuh (di luar jam 08:00 - 22:00)
	req := &models.CreateReservationRequest{
		CustomerName:    "Pelanggan Subuh",
		CustomerPhone:   "081234567890",
		ReservationDate: "2026-10-27",
		ReservationTime: "03:00",
		GuestCount:      2,
	}

	_, err := svc.CreatePublicReservation(context.Background(), req, "127.0.0.1")
	if err != ErrOutsideHours {
		t.Fatalf("Ekspektasi ErrOutsideHours untuk jam 03:00, didapat: %v", err)
	}
}

func TestTimeOverlapLogic(t *testing.T) {
	// Formula: newStart < existEnd && newEnd > existStart
	type slot struct {
		startMinutes int
		endMinutes   int
	}

	isOverlap := func(a, b slot) bool {
		return a.startMinutes < b.endMinutes && a.endMinutes > b.startMinutes
	}

	existing := slot{startMinutes: 14 * 60, endMinutes: 16 * 60} // 14:00 - 16:00 (120m)

	// Kasus 1: Mulai jam 15:00 - 17:00 (Bentrok di tengah)
	case1 := slot{startMinutes: 15 * 60, endMinutes: 17 * 60}
	if !isOverlap(case1, existing) {
		t.Error("Kasus 1: 15:00-17:00 seharusnya terdeteksi bentrok dengan 14:00-16:00")
	}

	// Kasus 2: Mulai jam 13:00 - 15:00 (Bentrok awal)
	case2 := slot{startMinutes: 13 * 60, endMinutes: 15 * 60}
	if !isOverlap(case2, existing) {
		t.Error("Kasus 2: 13:00-15:00 seharusnya terdeteksi bentrok dengan 14:00-16:00")
	}

	// Kasus 3: Mulai jam 16:00 - 18:00 (Tepat setelah slot selesai -> TIDAK bentrok)
	case3 := slot{startMinutes: 16 * 60, endMinutes: 18 * 60}
	if isOverlap(case3, existing) {
		t.Error("Kasus 3: 16:00-18:00 seharusnya TIDAK bentrok dengan 14:00-16:00")
	}

	// Kasus 4: Mulai jam 11:00 - 13:00 (Jauh sebelum -> TIDAK bentrok)
	case4 := slot{startMinutes: 11 * 60, endMinutes: 13 * 60}
	if isOverlap(case4, existing) {
		t.Error("Kasus 4: 11:00-13:00 seharusnya TIDAK bentrok dengan 14:00-16:00")
	}
}
