# KopiRise — Cafe Management System

Sistem manajemen operasional kafe & resto full-stack yang modern, aman, ringan, dan siap pakai untuk operasional **KopiRise Cafe & Resto** (Jl. Semeru, Kloncing, Karangrejo, Kec. Sumbersari, Kabupaten Jember, Jawa Timur 68124).

Mengintegrasikan website publik, katalog menu dinamis, reservasi meja anti bentrok, pemesanan mandiri melalui QR code meja, aplikasi kasir (POS), antrean dapur (Kitchen Display System), manajemen shift kas, inventaris bahan baku dengan Bill of Materials (BOM), kalkulasi HPP otomatis, serta pelaporan keuangan laba rugi komprehensif dalam satu kesatuan sistem.

---

## 🌟 Fitur Utama Sistem

1. **Website Publik & Profil Kafe**:
   - Beranda interaktif dengan galeri suasana 4 area kafe (Indoor AC, Outdoor Garden, Mezzanine, VIP Room Semeru).
   - Profil biji kopi spesialti lokal dari lereng Gunung Ijen & Raung.
   - Jam operasional terkonfigurasi (08:00 - 22:00 WIB) dan petunjuk arah Google Maps.

2. **Katalog Menu & Kustomisasi Dinamis**:
   - 32 item menu dalam 6 kategori makanan & minuman.
   - Varian ukuran dan modifier fleksibel (suhu panas/dingin, level gula, topping, ekstra espresso).
   - *Snapshotting* harga atomik: riwayat transaksi masa lalu tidak pernah berubah meski harga menu diubah di masa depan.

3. **Reservasi Meja Online**:
   - Pemilihan area & kapasitas tamu.
   - Algoritma pencegahan bentrok jadwal (*double-booking overlap prevention*).
   - Perlindungan anti-spam honeypot.
   - Kode reservasi unik (`RSV-YYYYMMDD-XXXX`) dan pelacakan status real-time.

4. **Pemesanan Mandiri QR Meja (Self-Service Ordering)**:
   - Pelanggan cukup memindai stiker QR pada meja (`/meja/[token]`).
   - Tidak memerlukan registrasi atau instalasi aplikasi native.
   - Pesanan langsung terdistribusi serentak ke layar Barista KDS dan POS Kasir.

5. **Aplikasi Kasir / POS & Cetak Struk Thermal**:
   - Buka dan tutup shift kasir dengan rekonsiliasi kas laci fisik vs sistem.
   - Mode Dine-In dan Takeaway.
   - Kalkulasi pembayaran tunai & kembalian instan, QRIS, dan transfer bank.
   - Cetak struk belanja thermal 58mm & 80mm lengkap dengan informasi Wi-Fi kafe.

6. **Kitchen Display System (KDS)**:
   - Antrean pesanan dapur & barista real-time dengan timer waktu tunggu.
   - Siklus progres: `Pending` $\rightarrow$ `Preparing` $\rightarrow$ `Ready to Serve` $\rightarrow$ `Served`.

7. **Keuangan, Inventaris Bahan Baku, & HPP**:
   - Bagan akun finansial (Kasir POS, Brankas Toko, Bank BCA, QRIS Mandiri) & mutasi antar-rekening.
   - Resep Bill of Materials (BOM) & kalkulasi HPP otomatis per porsi menu.
   - Pengurangan stok bahan baku otomatis saat hidangan selesai dimasak/terjual.
   - Pembelian bahan baku ke supplier dengan perhitungan biaya rata-rata tertimbang (*Weighted Average Cost*).
   - Laporan Laba Rugi Komprehensif dengan ekspor format Excel murni (`.xlsx`) dan CSV.

---

## 🛠️ Tech Stack

### Backend
- **Bahasa**: Go 1.22+
- **Router**: `github.com/go-chi/chi/v5`
- **Database Driver**: `modernc.org/sqlite` (Pure Go tanpa dependensi CGO)
- **Hashing**: Argon2id (`golang.org/x/crypto/argon2`)
- **Validasi**: `github.com/go-playground/validator/v10`
- **Rate Limiting**: `github.com/go-chi/httprate`
- **Migrasi**: `github.com/pressly/goose/v3`
- **Ekspor Excel**: `github.com/xuri/excelize/v2`

### Frontend
- **Framework**: SvelteKit dengan **Svelte 5 Runes** (`$state`, `$derived`, `$props`, `$effect`)
- **Bahasa**: TypeScript (Strict Mode)
- **Styling**: Tailwind CSS
- **Adapter**: SvelteKit `adapter-node`
- **Validasi Form**: Zod
- **Testing**: Playwright (E2E) & Vitest (Unit)

### Database & Infrastruktur
- **Basis Data**: SQLite murni dalam mode **Write-Ahead Logging (WAL)**
- **Backup**: Online hot-backup melalui `VACUUM INTO`
- **Container**: Docker Multi-Stage Build & Docker Compose
- **Reverse Proxy**: Caddy Server (Auto-HTTPS SSL)

---

## 🚀 Panduan Menjalankan Sistem

### Opsi 1: Menjalankan Langsung via Docker Compose (Rekomendasi Produksi)
```bash
# 1. Kloning repositori
git clone https://github.com/your-org/kopirise.git
cd kopirise

# 2. Siapkan file konfigurasi environment
cp .env.example .env

# 3. Jalankan seluruh container (Backend, Frontend, Caddy)
docker compose up -d --build
```
Aplikasi langsung dapat diakses di `http://localhost` atau domain yang dikonfigurasi.

### Opsi 2: Pengujian Langsung di Browser (Showcase Tanpa Server)
Buka file `d:\Kopirise\index.html` langsung di browser Chrome/Firefox/Edge. Seluruh modul (Web Publik, Menu, Reservasi, POS Kasir, Barista KDS, Manajemen Meja, Keuangan, dan SQLite Explorer) telah disematkan secara interaktif.

---

## 📂 Struktur Direktori Proyek

```text
d:\Kopirise\
├── backend/
│   ├── cmd/server/main.go            # Entrypoint HTTP server Chi & graceful shutdown
│   ├── internal/
│   │   ├── config/config.go          # Runtime environment configuration
│   │   ├── database/sqlite.go        # SQLite connection pool & PRAGMA WAL
│   │   ├── handler/                  # HTTP Handlers (Auth, Menu, Order, Payment, Finance, Inventory)
│   │   ├── middleware/               # Auth, RBAC, CORS, RequestID, Slog, Panic Recoverer
│   │   ├── models/models.go          # Domain structs & DTOs
│   │   ├── pkg/                      # Argon2id hashing & upload magic bytes validator
│   │   ├── repository/               # SQL transactional queries (User, Menu, Order, Payment, Finance, Inventory)
│   │   └── service/                  # Business logic (Auth, Order, Payment, Finance, Inventory, Export)
│   ├── migrations/                   # Goose migrations (00001 initial schema, 00002 seed dev)
│   └── scripts/                      # Automated test scripts (test_migrations.py, test_order_pos.py, etc.)
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/client.ts         # Typed fetch API client
│   │   │   ├── components/           # Svelte 5 Rune components (Navbar, Footer, MenuCard, ModifierModal, CartDrawer)
│   │   │   ├── types/index.ts        # TypeScript interfaces
│   │   │   └── utils/format.ts       # Format Rupiah & penanggalan WIB
│   │   ├── routes/
│   │   │   ├── +page.svelte          # Beranda Publik
│   │   │   ├── menu/                 # Katalog Menu Publik
│   │   │   ├── reservasi/            # Form Reservasi Online & Cek Status
│   │   │   ├── meja/[token]/         # Pemesanan Mandiri Meja QR
│   │   │   └── admin/                # POS Kasir, KDS Barista, Meja, Menu, & Laporan Keuangan
│   │   ├── app.html
│   │   ├── app.css                   # Tailwind CSS styling
│   │   └── hooks.server.ts           # Server-side auth & RBAC route guard
│   ├── tests/e2e/                    # Playwright E2E tests
│   ├── package.json
│   ├── svelte.config.js
│   ├── tailwind.config.js
│   └── vite.config.ts
├── docs/
│   ├── ARCHITECTURE.md               # Arsitektur sistem & prinsip desain
│   ├── ERD.md                        # Entity Relationship Diagram 41 tabel SQLite
│   ├── FINANCIAL_AND_TIME_RULES.md   # Ketentuan uang INTEGER Rupiah & waktu Asia/Jakarta
│   ├── DEPLOYMENT.md                 # Panduan deployment produksi, Caddy, & backup
│   └── OPERATIONAL_MANUAL.md         # SOP operasional staf kafe
├── data/                             # Direktori database SQLite (kopirise.db) & uploads
├── Caddyfile                         # Konfigurasi reverse proxy Caddy Auto-HTTPS
├── Dockerfile.backend                # Container Go production
├── Dockerfile.frontend               # Container SvelteKit Node 20
├── docker-compose.yml                # Orkestrasi multi-container
├── index.html                        # Showcase UI interaktif mandiri
└── README.md
```

---

## 🔒 Keamanan & Akurasi Finansial

- **Uang Tanpa Floating Point**: Seluruh nilai transaksi, harga, pembayaran, saldo, dan HPP disimpan murni sebagai `INTEGER` Rupiah.
- **Waktu Terstandarisasi**: Disimpan dalam database sebagai TEXT ISO 8601 UTC, dan dikonversikan ke zona waktu `Asia/Jakarta` (WIB) pada lapisan aplikasi.
- **Argon2id Password Hashing**: Dilengkapi pengujian timing-safe verification untuk mencegah serangan brute force dan timing attack.
- **Proteksi Idempotency**: Setiap pemesanan dan pembayaran kasir menyertakan kunci idempoten untuk mencegah pencatatan transaksi ganda.

---
© 2026 KopiRise Cafe & Resto Jember.
"# Tes-Cafe" 
