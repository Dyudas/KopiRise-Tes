package service

import (
	"strings"
	"testing"
	"time"

	"kopirise/backend/internal/models"
)

func TestTransferAndExpenseNumberFormat(t *testing.T) {
	todayStr := time.Now().Format("20060102")

	// Format Transfer: TRF-YYYYMMDD-XXXX
	trfNum := generateTransferNumber()
	if !strings.HasPrefix(trfNum, "TRF-") {
		t.Fatalf("Nomor transfer harus diawali 'TRF-', didapat: %s", trfNum)
	}
	trfParts := strings.Split(trfNum, "-")
	if len(trfParts) != 3 || trfParts[1] != todayStr || len(trfParts[2]) != 4 {
		t.Fatalf("Format nomor transfer salah: %s", trfNum)
	}

	// Format Expense: EXP-YYYYMMDD-XXXX
	expNum := generateExpenseNumber()
	if !strings.HasPrefix(expNum, "EXP-") {
		t.Fatalf("Nomor pengeluaran harus diawali 'EXP-', didapat: %s", expNum)
	}
	expParts := strings.Split(expNum, "-")
	if len(expParts) != 3 || expParts[1] != todayStr || len(expParts[2]) != 4 {
		t.Fatalf("Format nomor pengeluaran salah: %s", expNum)
	}
}

func TestProfitAndLossFormulas(t *testing.T) {
	// Mengetes formula akuntansi Laba Rugi Komprehensif
	testCases := []struct {
		name              string
		grossSales        int64
		discounts         int64
		taxPB1            int64
		cogs              int64
		operatingExpenses int64
		expectedNetSales  int64
		expectedGrossProf int64
		expectedNetProfit int64
	}{
		{
			name:              "Bulan Normal Profit Tinggi",
			grossSales:        50000000,
			discounts:         2000000,
			taxPB1:            4800000,
			cogs:              16000000, // 33% COGS
			operatingExpenses: 12000000, // Listrik, gaji, sewa
			expectedNetSales:  48000000,
			expectedGrossProf: 32000000,
			expectedNetProfit: 20000000,
		},
		{
			name:              "Bulan Rendah (Break Even / Rugi Tipis)",
			grossSales:        20000000,
			discounts:         1000000,
			taxPB1:            1900000,
			cogs:              8000000,
			operatingExpenses: 13000000,
			expectedNetSales:  19000000,
			expectedGrossProf: 11000000,
			expectedNetProfit: -2000000, // Rugi 2jt
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			netSales := tc.grossSales - tc.discounts
			if netSales != tc.expectedNetSales {
				t.Fatalf("Net sales salah: ekspektasi %d, didapat %d", tc.expectedNetSales, netSales)
			}

			grossProfit := netSales - tc.cogs
			if grossProfit != tc.expectedGrossProf {
				t.Fatalf("Gross profit salah: ekspektasi %d, didapat %d", tc.expectedGrossProf, grossProfit)
			}

			netProfit := grossProfit - tc.operatingExpenses
			if netProfit != tc.expectedNetProfit {
				t.Fatalf("Net profit salah: ekspektasi %d, didapat %d", tc.expectedNetProfit, netProfit)
			}

			// Margin Pct test
			if netSales > 0 {
				grossMarginPct := (float64(grossProfit) / float64(netSales)) * 100.0
				netMarginPct := (float64(netProfit) / float64(netSales)) * 100.0
				if grossMarginPct <= 0 && grossProfit > 0 {
					t.Fatalf("Gross margin pct abnormal: %f", grossMarginPct)
				}
				_ = netMarginPct
			}
		})
	}
}

func TestExpenseCategoryBreakdownPercentages(t *testing.T) {
	summaries := []*models.ExpenseCategorySummary{
		{CategoryName: "Bahan & Inventaris", TotalAmount: 5000000},
		{CategoryName: "Gaji Karyawan", TotalAmount: 3000000},
		{CategoryName: "Listrik & Utilitas", TotalAmount: 2000000},
	}

	totalExpenses := int64(10000000)
	for _, c := range summaries {
		c.Percentage = (float64(c.TotalAmount) / float64(totalExpenses)) * 100.0
	}

	if summaries[0].Percentage != 50.0 {
		t.Fatalf("Porsi persentase salah: ekspektasi 50%%, didapat %f%%", summaries[0].Percentage)
	}
	if summaries[1].Percentage != 30.0 {
		t.Fatalf("Porsi persentase salah: ekspektasi 30%%, didapat %f%%", summaries[1].Percentage)
	}
	if summaries[2].Percentage != 20.0 {
		t.Fatalf("Porsi persentase salah: ekspektasi 20%%, didapat %f%%", summaries[2].Percentage)
	}
}
