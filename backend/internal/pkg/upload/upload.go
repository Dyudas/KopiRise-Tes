package upload

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrFileTooLarge    = errors.New("ukuran file melebihi batas maksimal 5 MB")
	ErrInvalidFileType = errors.New("tipe file tidak diizinkan; hanya mendukung format JPEG, PNG, dan WebP")
	ErrEmptyFile       = errors.New("file tidak boleh kosong")
)

const MaxImageSize = 5 * 1024 * 1024 // 5 MB

// FileResult menyimpan informasi berkas yang telah berhasil disimpan.
type FileResult struct {
	OriginalName string
	StoredName   string
	FilePath     string
	FileSize     int64
	MimeType     string
}

// SaveImage memvalidasi magic bytes, batas ukuran, dan menyimpan gambar dengan nama acak UUID.
func SaveImage(uploadDir string, originalName string, r io.Reader) (*FileResult, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori upload: %w", err)
	}

	// Baca 512 byte pertama untuk deteksi MIME tipe dan magic bytes
	headerBytes := make([]byte, 512)
	n, err := r.Read(headerBytes)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("gagal membaca header file: %w", err)
	}
	if n == 0 {
		return nil, ErrEmptyFile
	}

	mimeType := http.DetectContentType(headerBytes[:n])
	ext := ""

	// Validasi Magic Bytes ketat
	switch {
	case strings.HasPrefix(mimeType, "image/jpeg"):
		if n >= 3 && headerBytes[0] == 0xFF && headerBytes[1] == 0xD8 && headerBytes[2] == 0xFF {
			ext = ".jpg"
		}
	case strings.HasPrefix(mimeType, "image/png"):
		if n >= 8 && bytes.Equal(headerBytes[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
			ext = ".png"
		}
	case strings.HasPrefix(mimeType, "image/webp"):
		if n >= 12 && string(headerBytes[0:4]) == "RIFF" && string(headerBytes[8:12]) == "WEBP" {
			ext = ".webp"
		}
	}

	if ext == "" {
		return nil, ErrInvalidFileType
	}

	// Buat nama berkas acak UUIDv4
	storedName := uuid.NewString() + ext
	destPath := filepath.Join(uploadDir, storedName)

	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat berkas tujuan: %w", err)
	}
	defer destFile.Close()

	// Tulis kembali 512 byte awal
	if _, err := destFile.Write(headerBytes[:n]); err != nil {
		_ = os.Remove(destPath)
		return nil, fmt.Errorf("gagal menulis header berkas: %w", err)
	}

	// Tulis sisa berkas dengan batas MaxImageSize
	limitedReader := io.LimitReader(r, MaxImageSize-int64(n)+1)
	written, err := io.Copy(destFile, limitedReader)
	if err != nil {
		_ = os.Remove(destPath)
		return nil, fmt.Errorf("gagal menulis isi berkas: %w", err)
	}

	totalSize := int64(n) + written
	if totalSize > MaxImageSize {
		_ = os.Remove(destPath)
		return nil, ErrFileTooLarge
	}

	return &FileResult{
		OriginalName: originalName,
		StoredName:   storedName,
		FilePath:     destPath,
		FileSize:     totalSize,
		MimeType:     mimeType,
	}, nil
}
