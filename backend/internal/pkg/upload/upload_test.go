package upload

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveImageValidPNG(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kopirise_upload_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Minimal PNG Header
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
	r := bytes.NewReader(pngBytes)

	res, err := SaveImage(tempDir, "sample.png", r)
	if err != nil {
		t.Fatalf("Gagal menyimpan PNG valid: %v", err)
	}

	if filepath.Ext(res.StoredName) != ".png" {
		t.Errorf("Ekspektasi ekstensi .png, didapat %s", filepath.Ext(res.StoredName))
	}

	if _, err := os.Stat(res.FilePath); os.IsNotExist(err) {
		t.Errorf("Berkas tidak ditemukan di lokasi tujuan: %s", res.FilePath)
	}
}

func TestSaveImageRejectsExecutable(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kopirise_upload_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Executable / script dummy
	badBytes := []byte("MZ\x90\x00\x03\x00\x00\x00#!/bin/bash\necho hack")
	r := bytes.NewReader(badBytes)

	_, err = SaveImage(tempDir, "script.sh", r)
	if err != ErrInvalidFileType {
		t.Fatalf("Seharusnya menolak tipe file selain gambar, didapat: %v", err)
	}
}
