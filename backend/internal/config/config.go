package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config menyimpan konfigurasi runtime aplikasi KopiRise.
type Config struct {
	AppEnv               string
	Port                 string
	Host                 string
	DBPath               string
	SessionCookieName    string
	SessionSecret        string
	SessionExpiry        time.Duration
	CookieSecure         bool
	CookieDomain         string
	CORSAllowedOrigins   []string
	MaxRequestBodyBytes  int64
	RateLimitRequests    int
	RateLimitWindow      time.Duration
	Argon2Memory         uint32
	Argon2Iterations     uint32
	Argon2Parallelism    uint8
	Argon2SaltLength     uint32
	Argon2KeyLength      uint32
}

// Load memuat konfigurasi dari environment variable dengan nilai default aman.
func Load() *Config {
	appEnv := getEnv("APP_ENV", "development")
	isProd := strings.ToLower(appEnv) == "production"

	sessionExpiryHours := getEnvAsInt("SESSION_EXPIRY_HOURS", 24)
	cookieSecure := getEnvAsBool("COOKIE_SECURE", isProd)

	origins := strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return &Config{
		AppEnv:              appEnv,
		Port:                getEnv("PORT", "8080"),
		Host:                getEnv("HOST", "0.0.0.0"),
		DBPath:              getEnv("DB_PATH", "data/kopirise.db"),
		SessionCookieName:   getEnv("SESSION_COOKIE_NAME", "kopirise_session"),
		SessionSecret:       getEnv("SESSION_SECRET", "super-secret-kopirise-jember-2026-key-replace-in-production"),
		SessionExpiry:       time.Duration(sessionExpiryHours) * time.Hour,
		CookieSecure:        cookieSecure,
		CookieDomain:        getEnv("COOKIE_DOMAIN", ""),
		CORSAllowedOrigins:  origins,
		MaxRequestBodyBytes: int64(getEnvAsInt("MAX_REQUEST_BODY_BYTES", 10*1024*1024)), // 10MB
		RateLimitRequests:   getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:     time.Minute,
		Argon2Memory:        uint32(getEnvAsInt("ARGON2_MEMORY", 64*1024)), // 64 MB
		Argon2Iterations:    uint32(getEnvAsInt("ARGON2_ITERATIONS", 3)),
		Argon2Parallelism:   uint8(getEnvAsInt("ARGON2_PARALLELISM", 2)),
		Argon2SaltLength:    uint32(getEnvAsInt("ARGON2_SALT_LENGTH", 16)),
		Argon2KeyLength:     uint32(getEnvAsInt("ARGON2_KEY_LENGTH", 32)),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}
