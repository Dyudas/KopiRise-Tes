package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kopirise/backend/internal/models"
)

func TestRequireAuth(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handlerToTest := RequireAuth(dummyHandler)

	// Case 1: Unauthenticated request -> should return 401
	req1 := httptest.NewRequest("GET", "/api/v1/users", nil)
	rec1 := httptest.NewRecorder()
	handlerToTest.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("Ekspektasi status 401 Unauthorized, didapat %d", rec1.Code)
	}

	// Case 2: Authenticated request with user context -> should return 200
	req2 := httptest.NewRequest("GET", "/api/v1/users", nil)
	ctx := context.WithValue(req2.Context(), models.ContextKeyUser, &models.User{
		ID:    "usr-01",
		Roles: []string{models.RoleCashier},
	})
	rec2 := httptest.NewRecorder()
	handlerToTest.ServeHTTP(rec2, req2.WithContext(ctx))

	if rec2.Code != http.StatusOK {
		t.Errorf("Ekspektasi status 200 OK, didapat %d", rec2.Code)
	}
}

func TestRequireRole(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("AUTHORIZED"))
	})

	adminOnlyHandler := RequireRole(models.RoleAdmin)(dummyHandler)

	// User is Cashier -> Admin only endpoint should return 403 Forbidden
	req1 := httptest.NewRequest("POST", "/api/v1/users", nil)
	ctx1 := context.WithValue(req1.Context(), models.ContextKeyUser, &models.User{
		ID:    "usr-02",
		Roles: []string{models.RoleCashier},
	})
	rec1 := httptest.NewRecorder()
	adminOnlyHandler.ServeHTTP(rec1, req1.WithContext(ctx1))

	if rec1.Code != http.StatusForbidden {
		t.Errorf("Ekspektasi status 403 Forbidden untuk role kasir, didapat %d", rec1.Code)
	}

	// User is Admin -> Admin only endpoint should return 200 OK
	req2 := httptest.NewRequest("POST", "/api/v1/users", nil)
	ctx2 := context.WithValue(req2.Context(), models.ContextKeyUser, &models.User{
		ID:    "usr-01",
		Roles: []string{models.RoleAdmin},
	})
	rec2 := httptest.NewRecorder()
	adminOnlyHandler.ServeHTTP(rec2, req2.WithContext(ctx2))

	if rec2.Code != http.StatusOK {
		t.Errorf("Ekspektasi status 200 OK untuk role admin, didapat %d", rec2.Code)
	}
}
