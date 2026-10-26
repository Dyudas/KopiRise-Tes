package service

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
	"kopirise/backend/internal/models"
)

type ExportService struct{}

func NewExportService() *ExportService {
	return &ExportService{}
}

// ExportProfitLossExcel menghasilkan file Excel (.xlsx) dengan formatting akuntansi rapi.
func (s *ExportService) ExportProfitLossExcel(
	report *models.ProfitLossReport,
	accounts []*models.FinancialAccount,
	ingredients []*models.Ingredient,
) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// -------------------------------------------------------------
	// SHEET 1: LAPORAN LABA RUGI
	// -------------------------------------------------------------
	sheetPL := "Laba Rugi"
	f.SetSheetName("Sheet1", sheetPL)

	// Judul Header
	f.SetCellValue(sheetPL, "A1", "KOPIRISE CAFE & RESTO JEMBER")
	f.SetCellValue(sheetPL, "A2", "LAPORAN LABA RUGI KOMPREHENSIF")
	f.SetCellValue(sheetPL, "A3", fmt.Sprintf("Periode: %s s/d %s (Mata Uang: IDR / Rupiah)", report.PeriodStart, report.PeriodEnd))

	// Baris Data Laba Rugi
	row := 5
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "KOMPONEN KEUANGAN")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), "NILAI (RUPIAH)")
	f.SetCellValue(sheetPL, fmt.Sprintf("C%d", row), "PERSENTASE")

	row++
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "Penjualan Kotor (Gross Sales)")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), report.GrossSales)

	row++
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "Potongan & Promo (Discounts)")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), report.Discounts)

	row++
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "Penjualan Bersih (Net Sales)")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), report.NetSales)
	f.SetCellValue(sheetPL, fmt.Sprintf("C%d", row), "100.0%")

	row++
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "Pajak Restoran Terkumpul (PB1 10%)")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), report.TaxCollected)

	row++
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "Harga Pokok Penjualan (HPP / COGS)")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), report.COGS)

	row++
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "LABA KOTOR (GROSS PROFIT)")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), report.GrossProfit)
	f.SetCellValue(sheetPL, fmt.Sprintf("C%d", row), fmt.Sprintf("%.2f%%", report.GrossMarginPct))

	row++
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "Total Beban Operasional (Operating Expenses)")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), report.OperatingExpenses)

	row++
	f.SetCellValue(sheetPL, fmt.Sprintf("A%d", row), "LABA BERSIH OPERASIONAL (NET PROFIT)")
	f.SetCellValue(sheetPL, fmt.Sprintf("B%d", row), report.NetProfit)
	f.SetCellValue(sheetPL, fmt.Sprintf("C%d", row), fmt.Sprintf("%.2f%%", report.NetMarginPct))

	f.SetColWidth(sheetPL, "A", "A", 45)
	f.SetColWidth(sheetPL, "B", "B", 25)
	f.SetColWidth(sheetPL, "C", "C", 20)

	// -------------------------------------------------------------
	// SHEET 2: RINCIAN BEBAN OPERASIONAL
	// -------------------------------------------------------------
	sheetExp := "Beban Operasional"
	f.NewSheet(sheetExp)
	f.SetCellValue(sheetExp, "A1", "RINCIAN PENGELUARAN & BEBAN OPERASIONAL")
	f.SetCellValue(sheetExp, "A3", "Kategori Beban")
	f.SetCellValue(sheetExp, "B3", "Total Pengeluaran (Rp)")
	f.SetCellValue(sheetExp, "C3", "Porsi Beban (%)")

	expRow := 4
	for _, exp := range report.ExpenseBreakdown {
		f.SetCellValue(sheetExp, fmt.Sprintf("A%d", expRow), exp.CategoryName)
		f.SetCellValue(sheetExp, fmt.Sprintf("B%d", expRow), exp.TotalAmount)
		f.SetCellValue(sheetExp, fmt.Sprintf("C%d", expRow), fmt.Sprintf("%.2f%%", exp.Percentage))
		expRow++
	}
	f.SetColWidth(sheetExp, "A", "A", 35)
	f.SetColWidth(sheetExp, "B", "B", 25)
	f.SetColWidth(sheetExp, "C", "C", 20)

	// -------------------------------------------------------------
	// SHEET 3: SALDO AKUN KEUANGAN
	// -------------------------------------------------------------
	sheetAcc := "Saldo Kas & Bank"
	f.NewSheet(sheetAcc)
	f.SetCellValue(sheetAcc, "A1", "SALDO REKENING & KAS FISIK KAFE")
	f.SetCellValue(sheetAcc, "A3", "Nama Akun")
	f.SetCellValue(sheetAcc, "B3", "Tipe Akun")
	f.SetCellValue(sheetAcc, "C3", "Nomor Rekening")
	f.SetCellValue(sheetAcc, "D3", "Saldo Terkini (Rp)")

	accRow := 4
	var totalBalance int64
	for _, acc := range accounts {
		f.SetCellValue(sheetAcc, fmt.Sprintf("A%d", accRow), acc.Name)
		f.SetCellValue(sheetAcc, fmt.Sprintf("B%d", accRow), acc.AccountType)
		f.SetCellValue(sheetAcc, fmt.Sprintf("C%d", accRow), acc.AccountNumber)
		f.SetCellValue(sheetAcc, fmt.Sprintf("D%d", accRow), acc.CurrentBalance)
		totalBalance += acc.CurrentBalance
		accRow++
	}
	f.SetCellValue(sheetAcc, fmt.Sprintf("A%d", accRow), "TOTAL LIKUIDITAS")
	f.SetCellValue(sheetAcc, fmt.Sprintf("D%d", accRow), totalBalance)

	f.SetColWidth(sheetAcc, "A", "A", 35)
	f.SetColWidth(sheetAcc, "B", "B", 18)
	f.SetColWidth(sheetAcc, "C", "C", 25)
	f.SetColWidth(sheetAcc, "D", "D", 25)

	// -------------------------------------------------------------
	// SHEET 4: STATUS STOK & VALUASI PERSEDIAAN
	// -------------------------------------------------------------
	sheetIng := "Persediaan & HPP"
	f.NewSheet(sheetIng)
	f.SetCellValue(sheetIng, "A1", "VALUASI STOK BAHAN BAKU")
	f.SetCellValue(sheetIng, "A3", "Nama Bahan")
	f.SetCellValue(sheetIng, "B3", "Satuan")
	f.SetCellValue(sheetIng, "C3", "Stok Tersedia")
	f.SetCellValue(sheetIng, "D3", "Batas Minimum")
	f.SetCellValue(sheetIng, "E3", "Biaya Satuan (Rp)")
	f.SetCellValue(sheetIng, "F3", "Valuasi Stok (Rp)")

	ingRow := 4
	var totalValuation int64
	for _, ing := range ingredients {
		valuation := int64(ing.CurrentStock * float64(ing.CostPerUnit))
		f.SetCellValue(sheetIng, fmt.Sprintf("A%d", ingRow), ing.Name)
		f.SetCellValue(sheetIng, fmt.Sprintf("B%d", ingRow), ing.Unit)
		f.SetCellValue(sheetIng, fmt.Sprintf("C%d", ingRow), ing.CurrentStock)
		f.SetCellValue(sheetIng, fmt.Sprintf("D%d", ingRow), ing.MinimumStock)
		f.SetCellValue(sheetIng, fmt.Sprintf("E%d", ingRow), ing.CostPerUnit)
		f.SetCellValue(sheetIng, fmt.Sprintf("F%d", ingRow), valuation)
		totalValuation += valuation
		ingRow++
	}
	f.SetCellValue(sheetIng, fmt.Sprintf("A%d", ingRow), "TOTAL NILAI PERSEDIAAN")
	f.SetCellValue(sheetIng, fmt.Sprintf("F%d", ingRow), totalValuation)

	f.SetColWidth(sheetIng, "A", "A", 35)
	f.SetColWidth(sheetIng, "B", "B", 15)
	f.SetColWidth(sheetIng, "C", "F", 20)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("write excel error: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportProfitLossCSV menghasilkan CSV ringkas untuk integrasi software akuntansi eksternal.
func (s *ExportService) ExportProfitLossCSV(report *models.ProfitLossReport) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("Komponen,Nilai_Rupiah,Persentase\n")
	buf.WriteString(fmt.Sprintf("Penjualan Kotor,%d,\n", report.GrossSales))
	buf.WriteString(fmt.Sprintf("Diskon,%d,\n", report.Discounts))
	buf.WriteString(fmt.Sprintf("Penjualan Bersih,%d,100.0%%\n", report.NetSales))
	buf.WriteString(fmt.Sprintf("Pajak PB1 Terkumpul,%d,\n", report.TaxCollected))
	buf.WriteString(fmt.Sprintf("HPP / COGS,%d,\n", report.COGS))
	buf.WriteString(fmt.Sprintf("Laba Kotor,%d,%.2f%%\n", report.GrossProfit, report.GrossMarginPct))
	buf.WriteString(fmt.Sprintf("Beban Operasional,%d,\n", report.OperatingExpenses))
	buf.WriteString(fmt.Sprintf("Laba Bersih,%d,%.2f%%\n", report.NetProfit, report.NetMarginPct))

	return buf.Bytes(), nil
}
