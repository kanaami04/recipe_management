package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
