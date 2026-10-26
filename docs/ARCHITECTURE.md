# Ringkasan Arsitektur & Keputusan Teknis — KopiRise System

## 1. Ringkasan Arsitektur (Maksimal 10 Baris)
1. KopiRise mengadopsi arsitektur modular monorepo: backend REST API Go terstruktur (`handler -> service -> repository`) dan frontend SvelteKit (Svelte 5 runes).
2. Database menggunakan SQLite murni Go (`modernc.org/sqlite`) berkinerja tinggi dengan mode WAL, foreign keys aktif, dan timeout sibuk 5 detik.
3. Keamanan mengandalkan autentikasi sesi server (`HttpOnly`, `SameSite=Lax`), enkripsi password Argon2id, RBAC berlapis di Go, dan token QR acak.
4. Integritas transaksi finansial dijamin melalui snapshot harga/pajak pada pesanan, idempotency key, larangan mutasi histori, dan reversal ter-audit.
5. Seluruh nilai uang disimpan sebagai `INTEGER` Rupiah tanpa koma desimal; persentase pajak dan diskon dihitung dengan rounding deterministik.
6. Seluruh timestamp disimpan dalam format teks ISO 8601 standar UTC (`Z`) dan dikonversi ke zona waktu `Asia/Jakarta` (WIB) pada layer aplikasi.
7. Alur operasional menyatukan QR pemesanan mandiri, POS kasir, antrean KDS barista, manajemen shift kas, reservasi meja, dan HPP inventaris.
8. Manajemen koneksi SQLite memisahkan single-writer pool untuk serialisasi mutasi data dan multi-reader pool untuk read concurrency tinggi.
9. Prosedur pencadangan data otomatis menggunakan `VACUUM INTO` aman tanpa mengunci pembacaan, dilengkapi panduan failover Litestream.
10. Antarmuka mengusung identitas visual premium coffee shop (Espresso `#3B2A22`, Cream `#F4EBDD`, Caramel `#B87945`) yang responsif dan mobile-first.

---

## 2. Struktur Folder Monorepo KopiRise
```text
d:\Kopirise/
├── index.html                           # Interactive Web App & System Preview (Buka langsung di browser)
├── docker-compose.yml                   # Orkestrasi Docker (Backend Go, Frontend SvelteKit, Caddy Reverse Proxy)
├── Caddyfile                            # Konfigurasi Caddy web server & reverse proxy SSL
├── Makefile                             # Script otomatisasi build, run, test, migrate, seed
├── .env.example                         # Template variabel lingkungan
├── docs/                                # Dokumentasi sistem & arsitektur
│   ├── ARCHITECTURE.md                  # Arsitektur & keputusan teknis
│   ├── ERD.md                           # Diagram relasi database (Mermaid)
│   └── FINANCIAL_AND_TIME_RULES.md      # Aturan perhitungan uang, HPP, formula kas, & timezone
├── data/                                # Volume persisten database & upload
│   ├── kopirise.db                      # SQLite main database file (WAL mode)
│   ├── kopirise.db-wal
│   ├── kopirise.db-shm
│   └── uploads/                         # Direktori file media & foto menu
├── backend/                             # REST API Go
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── server/main.go               # Entrypoint server Go HTTP Chi
│   │   ├── seeder/main.go               # CLI seeder development
│   │   └── admin-cli/main.go            # CLI bootstrap admin
│   ├── migrations/                      # Goose migration SQL scripts
│   │   ├── 00001_initial_schema.sql     # Skema DDL 41 tabel SQLite
│   │   └── 00002_seed_dev_data.sql      # Seeder realistis KopiRise
│   ├── scripts/
│   │   ├── test_migrations.py           # Script verifikasi integritas migrasi
│   │   └── backup_sqlite.py             # Script backup VACUUM INTO otomatis
│   ├── internal/
│   │   ├── config/                      # Manajemen konfigurasi env
│   │   ├── database/                    # SQLite connection pool & PRAGMA setup
│   │   ├── middleware/                  # Auth session, RBAC, CSRF, rate limit, request logger
│   │   ├── models/                      # Struktur data domain & DTO
│   │   ├── repository/                  # Query database/sql terorganisasi
│   │   ├── service/                     # Business logic, transaksi atomik, HPP, shift
│   │   └── handler/                     # HTTP Handler Chi, validator v10, response mapper
│   └── pkg/
│       ├── argon2id/                    # Hashing password argon2id
│       ├── excel/                       # Ekspor laporan Excel (excelize)
│       └── pdf/                         # Generator struk & invoice PDF
└── frontend/                            # Aplikasi SvelteKit (Svelte 5 Runes)
    ├── package.json
    ├── svelte.config.js
    ├── vite.config.ts
    ├── tailwind.config.cjs
    └── src/
        ├── app.html
        ├── lib/
        │   ├── components/              # UI components (Modal, Button, Table, Badge, Cart, Struk)
        │   ├── server/                  # Server-side auth, session, API client
        │   ├── stores/                  # Reactive state (Svelte 5 runes)
        │   └── types/                   # TypeScript interfaces API DTO
        └── routes/
            ├── (public)/                # Halaman publik (Beranda, Menu, Reservasi, Status Order)
            ├── qr/[token]/              # Halaman pesan mandiri QR Code Meja
            └── (admin)/                 # Halaman terproteksi
                ├── pos/                 # Aplikasi Kasir POS & Cetak Struk
                ├── kitchen/             # KDS Antrean Barista & Dapur
                ├── reservations/        # Manajemen kalender reservasi
                ├── tables/              # Manajemen Meja & Cetak QR
                ├── menu/                # Manajemen Menu, Varian, Modifier
                ├── finance/             # Pengeluaran, Kas, dan Laporan
                └── settings/            # Pengaturan toko & audit log
```

---

## 3. Batasan Konkurensi SQLite & Strategi Skalabilitas
- **Karakteristik SQLite**: Sangat cepat (mikrodetik) untuk pembacaan bersamaan (*concurrent reads*) berkat WAL (*Write-Ahead Logging*).
- **Batasan Tulis**: SQLite memiliki satu writer aktif pada satu saat. Untuk kafe dengan volume 500–2.000 pesanan per hari, waktu tulis per transaksi hanya 1–5 ms, sehingga kapasitas teoritis SQLite mampu menangani lebih dari 200–500 transaksi per detik tanpa lag.
- **Konfigurasi Proteksi Concurrency**:
  - `PRAGMA busy_timeout = 5000`: Mencegah error `SQLITE_BUSY` dengan menunggu antrean tulis hingga 5 detik.
  - Pool koneksi Go dibatasi menjadi `MaxOpenConns(1)` untuk instance writer database, atau menggunakan mutex transaction pool agar tidak terjadi race condition internal driver.
- **Kapan Harus Migrasi?**:
  - Jika kafe berkembang menjadi jaringan multi-cabang terdistribusi secara sinkron (*multi-region active-active write*), maka database relasional client-server seperti PostgreSQL dapat diadopsi dengan mudah karena arsitektur Go repository terisolasi rapi.
