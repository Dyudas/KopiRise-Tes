package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"kopirise/backend/internal/config"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/service"
)

// ResponseWriter wrapper untuk menangkap status HTTP response
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestID menyematkan UUID unik pada setiap request HTTP untuk audit tracing.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), models.ContextKeyRequestID, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StructuredLogger mencatat detail setiap request HTTP menggunakan log/slog tanpa mencatat data rahasia.
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		latency := time.Since(start)
		reqID, _ := r.Context().Value(models.ContextKeyRequestID).(string)

		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapper.statusCode,
			"duration_ms", latency.Milliseconds(),
			"ip", getClientIP(r),
			"user_agent", r.UserAgent(),
			"request_id", reqID,
		)
	})
}

// Recoverer menangkap panic runtime dan mengembalikan respon JSON 500 aman tanpa mengekspos stack trace ke klien.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID, _ := r.Context().Value(models.ContextKeyRequestID).(string)
				slog.Error("PANIC Terjadi pada Handler",
					"error", rec,
					"stack", string(debug.Stack()),
					"request_id", reqID,
					"path", r.URL.Path,
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(models.APIResponse{
					Success: false,
					Error: &models.APIErrorDetail{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Terjadi kesalahan internal pada server kami. Silakan coba kembali beberapa saat lagi.",
					},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS mengonfigurasi header Cross-Origin Resource Sharing yang aman.
func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			for _, o := range cfg.CORSAllowedOrigins {
				if o == "*" || strings.EqualFold(o, origin) {
					allowed = true
					break
				}
			}

			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BodySizeLimiter membatasi ukuran request body untuk mencegah DoS exhaustion attack.
func BodySizeLimiter(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// SessionAuth membaca session cookie dan melampirkan akun pengguna ke context request jika valid.
func SessionAuth(cfg *config.Config, authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.SessionCookieName)
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, err := authService.GetUserFromSession(r.Context(), cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), models.ContextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth menolak request yang tidak memiliki sesi autentikasi valid (401 Unauthorized).
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(models.ContextKeyUser).(*models.User)
		if !ok || user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(models.APIResponse{
				Success: false,
				Error: &models.APIErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "Akses ditolak: Silakan login terlebih dahulu untuk mengakses layanan ini.",
				},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole menegakkan izin Role-Based Access Control (RBAC) pada layer service/middleware backend (403 Forbidden).
func RequireRole(role string) func(http.Handler) http.Handler {
	return RequireAnyRole(role)
}

// RequireAnyRole mengizinkan request jika pengguna memiliki setidaknya salah satu peran yang dibutuhkan.
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(models.ContextKeyUser).(*models.User)
			if !ok || user == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(models.APIResponse{
					Success: false,
					Error: &models.APIErrorDetail{
						Code:    "UNAUTHORIZED",
						Message: "Silakan login terlebih dahulu.",
					},
				})
				return
			}

			if !user.HasAnyRole(roles...) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(models.APIResponse{
					Success: false,
					Error: &models.APIErrorDetail{
						Code:    "FORBIDDEN",
						Message: "Hak akses tidak mencukupi untuk modul atau tindakan ini.",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper untuk mendapatkan IP Client
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	return r.RemoteAddr
}
