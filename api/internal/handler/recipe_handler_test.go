package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
	appmw "recipe-backend/internal/middleware"
	jwtpkg "recipe-backend/internal/pkg/jwt"

	"github.com/labstack/echo/v4"
)

const testUserID = uint(7)

// bearerFor は testUserID 用のアクセストークンを付けたリクエストを作る。
func authedRequest(t *testing.T, jm *jwtpkg.Manager, method, path, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	token, err := jm.GenerateAccess(testUserID)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	return req
}

func TestRecipeHandler_List_Success(t *testing.T) {
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{
		listFn: func(_ context.Context, userID uint) ([]domain.Recipe, error) {
			if userID != testUserID {
				t.Errorf("userID = %d, want %d (from JWT)", userID, testUserID)
			}
			return []domain.Recipe{
				{ID: 1, Title: "肉じゃが", Owner: domain.ApplicationUser{ID: testUserID, Username: "alice"}},
			}, nil
		},
	})
	e.GET("/api/recipes/", h.List, appmw.JWT(jm))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodGet, "/api/recipes/", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "肉じゃが") {
		t.Errorf("body missing recipe: %s", rec.Body.String())
	}
}

func TestRecipeHandler_List_Unauthorized(t *testing.T) {
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{
		listFn: func(_ context.Context, _ uint) ([]domain.Recipe, error) {
			t.Fatal("service must not be called without auth")
			return nil, nil
		},
	})
	e.GET("/api/recipes/", h.List, appmw.JWT(jm))

	req := httptest.NewRequest(http.MethodGet, "/api/recipes/", nil) // トークンなし
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRecipeHandler_Create_Success(t *testing.T) {
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{
		createFn: func(_ context.Context, userID uint, req request.RecipeRequest) (*domain.Recipe, error) {
			if userID != testUserID {
				t.Errorf("userID = %d, want %d", userID, testUserID)
			}
			if req.Title != "カレー" {
				t.Errorf("title = %q, want カレー", req.Title)
			}
			return &domain.Recipe{ID: 9, Title: req.Title, Owner: domain.ApplicationUser{ID: testUserID, Username: "alice"}}, nil
		},
	})
	e.POST("/api/recipes/", h.Create, appmw.JWT(jm))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPost, "/api/recipes/", `{"title":"カレー"}`))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
}

func TestRecipeHandler_Create_ValidationError(t *testing.T) {
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{
		createFn: func(_ context.Context, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
			t.Fatal("service should not be called when validation fails")
			return nil, nil
		},
	})
	e.POST("/api/recipes/", h.Create, appmw.JWT(jm))

	// title 欠落 → validate:"required" で 400
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPost, "/api/recipes/", `{}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
