package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB membungkus *sql.DB dengan helper dan konfigurasi SQLite yang optimal.
type DB struct {
	*sql.DB
}

// Connect menginisialisasi koneksi SQLite dengan konfigurasi PRAGMA WAL dan concurrency yang aman.
func Connect(dbPath string) (*DB, error) {
	// Pastikan direktori database ada
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori database %s: %w", dir, err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi sqlite: %w", err)
	}

	// Batasi koneksi menjadi 1 untuk mencegah race condition locking SQLite pada operasi tulis
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("gagal ping database: %w", err)
	}

	// Pastikan pragma diterapkan secara eksplisit
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = NORMAL;",
	}

	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			slog.Warn("Gagal mengeksekusi PRAGMA", "pragma", p, "error", err)
		}
	}

	slog.Info("Koneksi SQLite berhasil diinisialisasi",
		"path", dbPath,
		"journal_mode", "WAL",
		"foreign_keys", "ON",
	)

	return &DB{DB: db}, nil
}

// Backup membuat cadangan instan menggunakan SQLite VACUUM INTO tanpa mengunci pembacaan.
func (db *DB) Backup(ctx context.Context, destPath string) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori backup: %w", err)
	}

	// Jika file tujuan sudah ada, hapus terlebih dahulu karena VACUUM INTO mengharuskan target baru
	if _, err := os.Stat(destPath); err == nil {
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("gagal menghapus file backup lama: %w", err)
		}
	}

	query := fmt.Sprintf("VACUUM INTO '%s';", destPath)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("eksekusi VACUUM INTO gagal: %w", err)
	}

	slog.Info("Database backup berhasil dibuat", "destination", destPath)
	return nil
}
