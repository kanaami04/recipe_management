package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
)

func TestAuthHandler_Token_Success(t *testing.T) {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{
		loginFn: func(_ context.Context, u, p string) (string, string, error) {
			return "access-tok", "refresh-tok", nil
		},
	})
	e.POST("/api/token/", h.Token)

	req := httptest.NewRequest(http.MethodPost, "/api/token/", strings.NewReader(`{"username":"alice","password":"pw"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "access-tok") || !strings.Contains(rec.Body.String(), "refresh-tok") {
		t.Errorf("body missing tokens: %s", rec.Body.String())
	}
}

func TestAuthHandler_Token_InvalidCredentials(t *testing.T) {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{
		loginFn: func(_ context.Context, u, p string) (string, string, error) {
			return "", "", service.ErrInvalidCredentials
		},
	})
	e.POST("/api/token/", h.Token)

	req := httptest.NewRequest(http.MethodPost, "/api/token/", strings.NewReader(`{"username":"alice","password":"bad"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestAuthHandler_Token_ValidationError(t *testing.T) {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{
		loginFn: func(_ context.Context, u, p string) (string, string, error) {
			t.Fatal("service should not be called when validation fails")
			return "", "", nil
		},
	})
	e.POST("/api/token/", h.Token)

	// password 欠落 → validate:"required" で 400
	req := httptest.NewRequest(http.MethodPost, "/api/token/", strings.NewReader(`{"username":"alice"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{
		registerFn: func(_ context.Context, u, em, p string) (*domain.ApplicationUser, error) {
			return &domain.ApplicationUser{ID: 1, Username: u, Email: em}, nil
		},
	})
	e.POST("/api/auth/register/", h.Register)

	body := `{"username":"alice","email":"alice@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "alice") {
		t.Errorf("body missing user: %s", rec.Body.String())
	}
}

func TestAuthHandler_Register_Conflict(t *testing.T) {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{
		registerFn: func(_ context.Context, u, em, p string) (*domain.ApplicationUser, error) {
			return nil, service.ErrUserAlreadyExists
		},
	})
	e.POST("/api/auth/register/", h.Register)

	body := `{"username":"alice","email":"alice@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{
		registerFn: func(_ context.Context, u, em, p string) (*domain.ApplicationUser, error) {
			t.Fatal("service should not be called when validation fails")
			return nil, nil
		},
	})
	e.POST("/api/auth/register/", h.Register)

	// email 不正 + password 8文字未満 → 400
	body := `{"username":"alice","email":"not-an-email","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAuthHandler_Token_MalformedBody(t *testing.T) {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{
		loginFn: func(_ context.Context, u, p string) (string, string, error) { return "", "", nil },
	})
	e.POST("/api/token/", h.Token)

	req := httptest.NewRequest(http.MethodPost, "/api/token/", strings.NewReader(`{not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
