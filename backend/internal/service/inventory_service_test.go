package service

import (
	"strings"
	"testing"
	"time"
)

func TestPurchaseNumberFormat(t *testing.T) {
	todayStr := time.Now().Format("20060102")

	poNum := generatePurchaseNumber()
	if !strings.HasPrefix(poNum, "PO-") {
		t.Fatalf("Nomor purchase order harus diawali 'PO-', didapat: %s", poNum)
	}

	parts := strings.Split(poNum, "-")
	if len(parts) != 3 || parts[1] != todayStr || len(parts[2]) != 4 {
		t.Fatalf("Format nomor PO salah: %s", poNum)
	}
}

func TestWeightedAverageCostFormula(t *testing.T) {
	// Formula Biaya Rata-Rata Tertimbang:
	// new_cost = ((old_stock * old_cost) + (purchase_qty * unit_cost)) / (old_stock + purchase_qty)
	testCases := []struct {
		name         string
		oldStock     float64
		oldCost      int64
		purchaseQty  float64
		unitCost     int64
		expectedCost int64
	}{
		{
			name:         "Pembelian Pertama Kali (Stok Awal Kosong)",
			oldStock:     0,
			oldCost:      0,
			purchaseQty:  10, // 10 kg
			unitCost:     120000,
			expectedCost: 120000,
		},
		{
			name:         "Pembelian Kedua Harga Naik (Kenaikan Harga Pasar)",
			oldStock:     5,      // sisa 5 kg @120.000 (total nilai 600.000)
			oldCost:      120000,
			purchaseQty:  5,      // beli lagi 5 kg @140.000 (total nilai 700.000)
			unitCost:     140000,
			expectedCost: 130000, // (600.000 + 700.000) / 10 = 1.300.000 / 10 = 130.000
		},
		{
			name:         "Pembelian Jumlah Lebih Besar Harga Diskon",
			oldStock:     2,      // 2 kg @130.000 = 260.000
			oldCost:      130000,
			purchaseQty:  8,      // 8 kg @100.000 = 800.000
			unitCost:     100000,
			expectedCost: 106000, // (260.000 + 800.000) / 10 = 1.060.000 / 10 = 106.000
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newStock := tc.oldStock + tc.purchaseQty
			var calculatedCost int64
			if newStock > 0 {
				totalValuation := (tc.oldStock * float64(tc.oldCost)) + (tc.purchaseQty * float64(tc.unitCost))
				calculatedCost = int64(totalValuation / newStock)
			} else {
				calculatedCost = tc.unitCost
			}

			if calculatedCost != tc.expectedCost {
				t.Fatalf("Biaya rata-rata tertimbang salah: ekspektasi %d, didapat %d", tc.expectedCost, calculatedCost)
			}
		})
	}
}

func TestRecipeBOMHPPAndMarginCalculation(t *testing.T) {
	// Mengetes kalkulasi HPP Bill of Materials Kopi Susu Aren Semeru
	type IngredientPortion struct {
		name        string
		qty         float64
		costPerUnit int64
	}

	sellingPrice := int64(22000)
	recipeItems := []IngredientPortion{
		{name: "Biji Kopi Arabika Raung (gram)", qty: 18.0, costPerUnit: 180},    // 18g * Rp 180 = Rp 3.240
		{name: "Susu Segar UHT Fresh Milk (ml)", qty: 150.0, costPerUnit: 24},   // 150ml * Rp 24 = Rp 3.600
		{name: "Sirup Gula Aren Asli (ml)", qty: 25.0, costPerUnit: 45},         // 25ml * Rp 45 = Rp 1.125
		{name: "Cup KopiRise & Straw Lid (pcs)", qty: 1.0, costPerUnit: 1250},   // 1pcs * Rp 1.250 = Rp 1.250
	}

	var calculatedHPP int64
	for _, itm := range recipeItems {
		itemCost := int64(itm.qty * float64(itm.costPerUnit))
		calculatedHPP += itemCost
	}

	expectedHPP := int64(9215)
	if calculatedHPP != expectedHPP {
		t.Fatalf("Total HPP salah: ekspektasi %d, didapat %d", expectedHPP, calculatedHPP)
	}

	marginRp := sellingPrice - calculatedHPP
	expectedMarginRp := int64(12785)
	if marginRp != expectedMarginRp {
		t.Fatalf("Margin Rp salah: ekspektasi %d, didapat %d", expectedMarginRp, marginRp)
	}

	marginPct := (float64(marginRp) / float64(sellingPrice)) * 100.0
	// Ekspektasi ~58.11%
	if marginPct < 58.0 || marginPct > 58.2 {
		t.Fatalf("Margin percentage di luar ekspektasi: %f%%", marginPct)
	}
}
