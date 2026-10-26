package service

import (
	"strings"
	"testing"
	"time"
)

func TestPaymentAndRefundNumberFormat(t *testing.T) {
	todayStr := time.Now().Format("20060102")

	// Format Pembayaran: PAY-YYYYMMDD-XXXX
	payNum := generatePaymentNumber()
	if !strings.HasPrefix(payNum, "PAY-") {
		t.Fatalf("Nomor pembayaran harus diawali 'PAY-', didapat: %s", payNum)
	}
	payParts := strings.Split(payNum, "-")
	if len(payParts) != 3 || payParts[1] != todayStr || len(payParts[2]) != 4 {
		t.Fatalf("Format nomor pembayaran salah: %s", payNum)
	}

	// Format Refund: REF-YYYYMMDD-XXXX
	refNum := generateRefundNumber()
	if !strings.HasPrefix(refNum, "REF-") {
		t.Fatalf("Nomor refund harus diawali 'REF-', didapat: %s", refNum)
	}
	refParts := strings.Split(refNum, "-")
	if len(refParts) != 3 || refParts[1] != todayStr || len(refParts[2]) != 4 {
		t.Fatalf("Format nomor refund salah: %s", refNum)
	}
}

func TestCashPaymentChangeCalculation(t *testing.T) {
	testCases := []struct {
		name           string
		billTotal      int64
		cashTendered   int64
		expectedChange int64
		expectValid    bool
	}{
		{
			name:           "Uang Pas Rp 75.000",
			billTotal:      75000,
			cashTendered:   75000,
			expectedChange: 0,
			expectValid:    true,
		},
		{
			name:           "Kembalian Rp 25.000 (Bayar 100rb untuk tagihan 75rb)",
			billTotal:      75000,
			cashTendered:   100000,
			expectedChange: 25000,
			expectValid:    true,
		},
		{
			name:           "Uang Kurang (Bayar 50rb untuk tagihan 75rb)",
			billTotal:      75000,
			cashTendered:   50000,
			expectedChange: 0,
			expectValid:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cashTendered < tc.billTotal {
				if tc.expectValid {
					t.Fatalf("Transaksi seharusnya tidak valid ketika uang kurang")
				}
				return
			}

			change := tc.cashTendered - tc.billTotal
			if change != tc.expectedChange {
				t.Fatalf("Kembalian salah: ekspektasi %d, didapat %d", tc.expectedChange, change)
			}
		})
	}
}

func TestCashShiftReconciliationFormula(t *testing.T) {
	// Formula: kas_seharusnya = kas_awal + penjualan_tunai + kas_masuk - kas_keluar - refund_tunai
	// selisih = kas_fisik_aktual - kas_seharusnya
	testCases := []struct {
		name             string
		startingCash     int64
		cashSales        int64
		cashIn           int64
		cashOut          int64
		refundCash       int64
		actualPhysical   int64
		expectedCash     int64
		expectedVariance int64
		status           string
	}{
		{
			name:             "Shift Seimbang (Sesuai)",
			startingCash:     500000,
			cashSales:        1250000,
			cashIn:           100000, // modal koin tambahan
			cashOut:          50000,  // beli gas lpg
			refundCash:       0,
			actualPhysical:   1800000,
			expectedCash:     1800000,
			expectedVariance: 0,
			status:           "balanced",
		},
		{
			name:             "Shift Kas Kurang (Minus)",
			startingCash:     500000,
			cashSales:        1000000,
			cashIn:           0,
			cashOut:          0,
			refundCash:       50000,  // refund tunai pelanggan
			actualPhysical:   1430000, // fisik di laci kurang 20rb
			expectedCash:     1450000,
			expectedVariance: -20000,
			status:           "shortage",
		},
		{
			name:             "Shift Kas Lebih (Surplus)",
			startingCash:     500000,
			cashSales:        800000,
			cashIn:           0,
			cashOut:          0,
			refundCash:       0,
			actualPhysical:   1315000, // lebih 15rb karena tip pelanggan masuk laci
			expectedCash:     1300000,
			expectedVariance: 15000,
			status:           "overage",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calcExpected := tc.startingCash + tc.cashSales + tc.cashIn - tc.cashOut - tc.refundCash
			if calcExpected != tc.expectedCash {
				t.Fatalf("Kas seharusnya salah: ekspektasi %d, didapat %d", tc.expectedCash, calcExpected)
			}

			calcVariance := tc.actualPhysical - calcExpected
			if calcVariance != tc.expectedVariance {
				t.Fatalf("Selisih kas salah: ekspektasi %d, didapat %d", tc.expectedVariance, calcVariance)
			}
		})
	}
}
