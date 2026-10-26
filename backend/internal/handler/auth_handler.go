package handler

import (
	"errors"
	"net/http"
	"strings"

	"kopirise/backend/internal/config"
	"kopirise/backend/internal/models"
	"kopirise/backend/internal/service"
)

// AuthHandler menangani endpoint HTTP autentikasi dan sesi.
type AuthHandler struct {
	cfg         *config.Config
	authService *service.AuthService
}

// NewAuthHandler membuat instance AuthHandler.
func NewAuthHandler(cfg *config.Config, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		cfg:         cfg,
		authService: authService,
	}
}

// Login memproses request masuk pengguna.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	ip := getIP(r)
	ua := r.UserAgent()

	user, session, err := h.authService.Login(r.Context(), req.Email, req.Password, ip, ua)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			RespondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email atau kata sandi yang Anda masukkan salah.", nil)
			return
		}
		if errors.Is(err, service.ErrAccountInactive) {
			RespondError(w, http.StatusForbidden, "ACCOUNT_INACTIVE", "Akun ini sedang dinonaktifkan atau ditangguhkan.", nil)
			return
		}
		RespondError(w, http.StatusInternalServerError, "LOGIN_FAILED", "Gagal memproses login.", nil)
		return
	}

	// Set Secure Session Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Domain:   h.cfg.CookieDomain,
	})

	RespondSuccess(w, models.LoginResponse{
		User:      user,
		ExpiresAt: session.ExpiresAt,
	})
}

// Logout mengakhiri sesi pengguna dan menghapus cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie(h.cfg.SessionCookieName)
	var sessionID string
	if cookie != nil {
		sessionID = cookie.Value
	}

	var userID *string
	if user, ok := r.Context().Value(models.ContextKeyUser).(*models.User); ok && user != nil {
		userID = &user.ID
	}

	_ = h.authService.Logout(r.Context(), sessionID, userID, getIP(r))

	// Hapus cookie di browser klien
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	RespondSuccess(w, map[string]string{
		"message": "Logout berhasil.",
	})
}

// Me mengembalikan informasi akun dan peran pengguna yang sedang login.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(models.ContextKeyUser).(*models.User)
	if !ok || user == nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tidak ada sesi pengguna aktif.", nil)
		return
	}

	RespondSuccess(w, user)
}

func getIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	return r.RemoteAddr
}
