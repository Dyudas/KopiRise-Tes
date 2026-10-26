# Aturan Perhitungan Keuangan, HPP, Formula Kas, & Format Waktu

## 1. Aturan Penyimpanan Nilai Uang (INTEGER Rupiah)
- **Larangan Float**: Seluruh kolom moneter pada database disimpan sebagai `INTEGER` dalam satuan terkecil mata uang Rupiah (IDR). Contoh:
  - Rp 22.000 disimpan sebagai integer `22000`.
  - Rp 3.500.000 disimpan sebagai integer `3500000`.
  - Tidak ada floating point (`FLOAT` / `REAL`) untuk uang guna menghindari akumulasi ketidaktepatan biner (*floating point rounding errors*).

## 2. Formula Perhitungan Transaksi & Keranjang Pesanan

```text
subtotal = SUM( (unit_base_price + SUM(modifier_price)) * quantity )

diskon = 
  - Jika tipe fixed_amount : MIN(promo.discount_value, subtotal)
  - Jika tipe percentage   : MIN( FLOOR(subtotal * (promo.discount_value / 100)), promo.max_discount_amount )

dpp = subtotal - diskon   (Dasar Pengenaan Pajak)

service_charge = FLOOR( dpp * (service_charge_percent / 100) )

pajak = ROUND( (dpp + service_charge) * (tax_rate_percent / 100) )

total_amount = dpp + service_charge + pajak
```

Contoh Kasus Riil:
- 1x Kopi Susu Aren Semeru (`22000`) + Extra Shot (`5000`) = `27000`
- 1x Tahu Walik Krispi Jember (`18000`) = `18000`
- **Subtotal**: `45000`
- Diskon Promo 15%: `FLOOR(45000 * 0.15)` = `6750`
- **DPP**: `45000 - 6750` = `38250`
- Pajak PB1 (10%): `ROUND(38250 * 0.10)` = `3825`
- **Total Pembayaran**: `38250 + 3825` = `42075` (dibulatkan sesuai kebijakan kasir atau diterima presisi Rp 42.075 via QRIS).

---

## 3. Formula Rekonsiliasi Kas Shift (POS)
Setiap kasir memulai shift dengan modal awal (*opening balance*) dan mencatat seluruh pergerakan uang tunai fisik.

$$\text{kas\_seharusnya} = \text{kas\_awal} + \text{penjualan\_tunai} + \text{kas\_masuk} - \text{kas\_keluar} - \text{refund\_tunai}$$

- $\text{kas\_awal}$: Modal kembalian awal di laci kasir saat buka shift.
- $\text{penjualan\_tunai}$: Akumulasi pembayaran berstatus `completed` dengan metode `cash` dalam rentang waktu shift.
- $\text{kas\_masuk}$: Tambahan modal operasional dari brankas/manager selama shift berlangsung.
- $\text{kas\_keluar}$: Pengeluaran tunai darurat (misal: beli es batu kristal tambahan) yang telah disetujui.
- $\text{refund\_tunai}$: Pengembalian dana tunai kepada pelanggan untuk transaksi yang dibatalkan sah.
- $\text{selisih\_kas} = \text{kas\_fisik\_aktual} - \text{kas\_seharusnya}$
  - Jika `selisih = 0`: Kas seimbang (Balance).
  - Jika `selisih > 0`: Selisih lebih (*overage*). Wajib dicatat keterangannya.
  - Jika `selisih < 0`: Selisih kurang (*shortage*). Tanggung jawab kasir bertugas.

---

## 4. Perhitungan HPP (Harga Pokok Penjualan) & Margin
- **HPP Per Porsi Menu**:
  $$\text{HPP}_{\text{menu}} = \sum (\text{kuantitas\_bahan} \times \text{biaya\_per\_satuan})$$
- Contoh: Kopi Susu Aren Semeru:
  - 18 gram Biji Kopi Arabika Ijen @ Rp 280 = Rp 5.040
  - 150 ml Fresh Milk @ Rp 22 = Rp 3.300
  - 25 ml Sirup Gula Aren @ Rp 35 = Rp 875
  - **Total HPP**: **Rp 9.215**
  - Harga Jual: **Rp 22.000**
  - **Laba Kotor Estimasi**: $\text{Rp } 22.000 - \text{Rp } 9.215 = \text{\textbf{Rp 12.785}}$ ($\text{Margin } 58{,}1\%$)

> **Catatan Penting Akuntansi**: Jika suatu menu belum didaftarkan resepnya pada tabel `recipes` atau harga bahan baku belum terisi, sistem akan menandai: *"Estimasi Margin Belum Lengkap"*, mencegah distorsi pada laporan laba operasional kafe.

---

## 5. Standar Waktu & Konversi Zona Waktu
1. **Penyimpanan Database**:
   - Seluruh kolom waktu (`created_at`, `updated_at`, `expires_at`, `opened_at`, dsb.) disimpan sebagai string standar **ISO 8601 UTC** dengan akhiran `Z`.
   - Format: `YYYY-MM-DDTHH:MM:SSZ` (contoh: `2026-10-26T03:49:53Z`).
2. **Presentasi Aplikasi**:
   - Lapisan server Go dan frontend SvelteKit selalu mengonversi timestamp UTC ke zona waktu **Asia/Jakarta** (WIB, UTC+7).
   - `2026-10-26T03:49:53Z` $\rightarrow$ `26 Oktober 2026, 10:49:53 WIB`.
3. **Kolom Tanggal Murni**:
   - Untuk tanggal kalender seperti tanggal shift atau tanggal reservasi disimpan dalam format ISO date lokal: `YYYY-MM-DD` (contoh: `2026-10-26`).
   - Waktu jam reservasi: `HH:MM:SS` (contoh: `14:00:00`).
