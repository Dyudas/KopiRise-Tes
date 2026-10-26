package argon2id

import (
	"testing"
)

func TestHashPasswordAndVerify(t *testing.T) {
	password := "KopiRiseJember2026!"

	// Gunakan parameter cepat untuk pengujian unit
	testParams := &Params{
		Memory:      16 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := HashPassword(password, testParams)
	if err != nil {
		t.Fatalf("Gagal membuat hash: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash tidak boleh kosong")
	}

	// Verifikasi kata sandi benar
	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("Verifikasi menghasilkan error: %v", err)
	}
	if !match {
		t.Fatal("Verifikasi seharusnya berhasil untuk password yang benar")
	}

	// Verifikasi kata sandi salah harus gagal
	matchWrong, err := VerifyPassword("PasswordSalah123", hash)
	if err != nil {
		t.Fatalf("Verifikasi password salah menghasilkan error tak terduga: %v", err)
	}
	if matchWrong {
		t.Fatal("Verifikasi seharusnya gagal untuk password yang salah")
	}
}

func TestVerifyInvalidHashFormat(t *testing.T) {
	_, err := VerifyPassword("test", "invalid-hash-string")
	if err == nil {
		t.Fatal("Seharusnya menghasilkan error untuk format hash yang tidak valid")
	}
}
