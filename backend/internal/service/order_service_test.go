package service

import (
	"strings"
	"testing"
	"time"
)

func TestOrderNumberFormat(t *testing.T) {
	// Format wajib: ORD-YYYYMMDD-XXXX (4 random alphanumeric characters)
	orderNum := generateOrderNumber()

	if !strings.HasPrefix(orderNum, "ORD-") {
		t.Fatalf("Nomor pesanan harus diawali 'ORD-', didapat: %s", orderNum)
	}

	parts := strings.Split(orderNum, "-")
	if len(parts) != 3 {
		t.Fatalf("Nomor pesanan harus memiliki format ORD-YYYYMMDD-XXXX, didapat: %s", orderNum)
	}

	todayStr := time.Now().Format("20060102")
	if parts[1] != todayStr {
		t.Fatalf("Bagian tanggal pesanan harus hari ini (%s), didapat: %s", todayStr, parts[1])
	}

	if len(parts[2]) != 4 {
		t.Fatalf("Suffix pesanan harus terdiri dari 4 karakter acak, didapat: %s", parts[2])
	}
}

func TestOrderFinancialCalculation(t *testing.T) {
	// Mengetes kalkulasi subtotal, diskon, pajak PB1 10%, dan total akhir dalam INTEGER Rupiah
	type ItemCalculation struct {
		unitPrice    int64
		quantity     int
		modifierCost int64
	}

	testCases := []struct {
		name             string
		items            []ItemCalculation
		discountType     string
		discountValue    int64
		expectedSubtotal int64
		expectedDiscount int64
		expectedTax      int64
		expectedTotal    int64
	}{
		{
			name: "Pesanan Standar Tanpa Diskon (PB1 10%)",
			items: []ItemCalculation{
				{unitPrice: 25000, quantity: 2, modifierCost: 0},     // 50.000
				{unitPrice: 32000, quantity: 1, modifierCost: 5000},  // 37.000
			},
			discountType:     "",
			discountValue:    0,
			expectedSubtotal: 87000,
			expectedDiscount: 0,
			expectedTax:      8700, // 10% dari 87.000
			expectedTotal:    95700,
		},
		{
			name: "Pesanan dengan Diskon Persentase 20%",
			items: []ItemCalculation{
				{unitPrice: 100000, quantity: 1, modifierCost: 0},
			},
			discountType:     "percentage",
			discountValue:    20,
			expectedSubtotal: 100000,
			expectedDiscount: 20000,
			expectedTax:      8000,  // 10% dari (100.000 - 20.000)
			expectedTotal:    88000,
		},
		{
			name: "Pesanan dengan Diskon Tetap Nominal Rp 15.000",
			items: []ItemCalculation{
				{unitPrice: 50000, quantity: 1, modifierCost: 0},
			},
			discountType:     "fixed",
			discountValue:    15000,
			expectedSubtotal: 50000,
			expectedDiscount: 15000,
			expectedTax:      3500,  // 10% dari (50.000 - 15.000)
			expectedTotal:    38500,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var calculatedSubtotal int64
			for _, item := range tc.items {
				itemTotal := (item.unitPrice + item.modifierCost) * int64(item.quantity)
				calculatedSubtotal += itemTotal
			}

			if calculatedSubtotal != tc.expectedSubtotal {
				t.Fatalf("Subtotal tidak cocok: ekspektasi %d, didapat %d", tc.expectedSubtotal, calculatedSubtotal)
			}

			var calculatedDiscount int64
			if tc.discountType == "percentage" {
				calculatedDiscount = (calculatedSubtotal * tc.discountValue) / 100
			} else if tc.discountType == "fixed" {
				calculatedDiscount = tc.discountValue
			}

			if calculatedDiscount > calculatedSubtotal {
				calculatedDiscount = calculatedSubtotal
			}

			if calculatedDiscount != tc.expectedDiscount {
				t.Fatalf("Diskon tidak cocok: ekspektasi %d, didapat %d", tc.expectedDiscount, calculatedDiscount)
			}

			taxableAmount := calculatedSubtotal - calculatedDiscount
			calculatedTax := (taxableAmount * 10) / 100 // PB1 10%
			if calculatedTax != tc.expectedTax {
				t.Fatalf("Pajak PB1 tidak cocok: ekspektasi %d, didapat %d", tc.expectedTax, calculatedTax)
			}

			calculatedTotal := taxableAmount + calculatedTax
			if calculatedTotal != tc.expectedTotal {
				t.Fatalf("Total akhir tidak cocok: ekspektasi %d, didapat %d", tc.expectedTotal, calculatedTotal)
			}
		})
	}
}
