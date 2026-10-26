package config

import (
	"os"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	os.Setenv("PORT", "9090")
	os.Setenv("DB_PATH", "data/test_kopirise.db")

	cfg := Load()

	if cfg.AppEnv != "test" {
		t.Errorf("Ekspektasi AppEnv 'test', didapat %s", cfg.AppEnv)
	}
	if cfg.Port != "9090" {
		t.Errorf("Ekspektasi Port '9090', didapat %s", cfg.Port)
	}
	if cfg.DBPath != "data/test_kopirise.db" {
		t.Errorf("Ekspektasi DBPath 'data/test_kopirise.db', didapat %s", cfg.DBPath)
	}
	if cfg.SessionCookieName != "kopirise_session" {
		t.Errorf("Ekspektasi Cookie Name 'kopirise_session', didapat %s", cfg.SessionCookieName)
	}
}
