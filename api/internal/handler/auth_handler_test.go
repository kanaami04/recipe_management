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
	"github.com/stretchr/testify/assert"
)

// serveToken は loginFn を差し替えた AuthHandler に POST /api/token/ し、結果を返す。
func serveToken(loginFn func(context.Context, string, string) (string, string, error), body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{loginFn: loginFn})
	e.POST("/api/token/", h.Token)
	req := httptest.NewRequest(http.MethodPost, "/api/token/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// serveRegister は registerFn を差し替えた AuthHandler に POST /api/auth/register/ し、結果を返す。
func serveRegister(registerFn func(context.Context, string, string, string) (*domain.ApplicationUser, error), body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{registerFn: registerFn})
	e.POST("/api/auth/register/", h.Register)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// okLogin は常に固定トークンを返す loginFn。
func okLogin(_ context.Context, _, _ string) (string, string, error) {
	return "access-tok", "refresh-tok", nil
}

// okRegister は受け取った username/email でユーザーを返す registerFn。
func okRegister(_ context.Context, u, em, _ string) (*domain.ApplicationUser, error) {
	return &domain.ApplicationUser{ID: 1, Username: u, Email: em}, nil
}

const (
	validLoginBody    = `{"username":"alice","password":"pw"}`
	validRegisterBody = `{"username":"alice","email":"alice@example.com","password":"password123"}`
	validRefreshBody  = `{"refresh":"some-refresh-token"}`
)

// serveRefresh は refreshFn を差し替えた AuthHandler に POST /api/token/refresh/ し、結果を返す。
func serveRefresh(refreshFn func(context.Context, string) (string, error), body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{refreshFn: refreshFn})
	e.POST("/api/token/refresh/", h.Refresh)
	req := httptest.NewRequest(http.MethodPost, "/api/token/refresh/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// okRefresh は固定のアクセストークンを返す refreshFn。
func okRefresh(_ context.Context, _ string) (string, error) {
	return "new-access-tok", nil
}

// 正しい資格情報でトークン取得した時、200 が返ること。
func TestAuthHandler_Token_Success_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveToken(okLogin, validLoginBody)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 正しい資格情報でトークン取得した時、レスポンスにアクセストークンが含まれること。
func TestAuthHandler_Token_Success_ReturnsAccessToken(t *testing.T) {
	// Arrange & Act
	rec := serveToken(okLogin, validLoginBody)

	// Assert
	assert.Contains(t, rec.Body.String(), "access-tok")
}

// 正しい資格情報でトークン取得した時、レスポンスにリフレッシュトークンが含まれること。
func TestAuthHandler_Token_Success_ReturnsRefreshToken(t *testing.T) {
	// Arrange & Act
	rec := serveToken(okLogin, validLoginBody)

	// Assert
	assert.Contains(t, rec.Body.String(), "refresh-tok")
}

// 認証情報が誤っている時、401 が返ること。
func TestAuthHandler_Token_InvalidCredentials(t *testing.T) {
	// Arrange & Act
	rec := serveToken(func(_ context.Context, _, _ string) (string, string, error) {
		return "", "", service.ErrInvalidCredentials
	}, `{"username":"alice","password":"bad"}`)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// 必須項目が欠けている時、サービスを呼ばず 400 が返ること。
func TestAuthHandler_Token_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveToken(func(_ context.Context, _, _ string) (string, string, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return "", "", nil
	}, `{"username":"alice"}`) // password 欠落

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// JSON として壊れたボディの時、400 が返ること。
func TestAuthHandler_Token_MalformedBody(t *testing.T) {
	// Arrange & Act
	rec := serveToken(func(_ context.Context, _, _ string) (string, string, error) {
		return "", "", nil
	}, `{not json`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 新規登録に成功した時、201 が返ること。
func TestAuthHandler_Register_Success_Returns201(t *testing.T) {
	// Arrange & Act
	rec := serveRegister(okRegister, validRegisterBody)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
}

// 新規登録に成功した時、レスポンスに登録ユーザーが含まれること。
func TestAuthHandler_Register_Success_ReturnsUser(t *testing.T) {
	// Arrange & Act
	rec := serveRegister(okRegister, validRegisterBody)

	// Assert
	assert.Contains(t, rec.Body.String(), "alice")
}

// 既存ユーザーと重複する時、409 が返ること。
func TestAuthHandler_Register_Conflict(t *testing.T) {
	// Arrange & Act
	rec := serveRegister(func(_ context.Context, _, _, _ string) (*domain.ApplicationUser, error) {
		return nil, service.ErrUserAlreadyExists
	}, validRegisterBody)

	// Assert
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// email 不正やパスワード不足の時、サービスを呼ばず 400 が返ること。
func TestAuthHandler_Register_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveRegister(func(_ context.Context, _, _, _ string) (*domain.ApplicationUser, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, `{"username":"alice","email":"not-an-email","password":"short"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 有効なリフレッシュトークンで更新した時、200 が返ること。
func TestAuthHandler_Refresh_Success_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(okRefresh, validRefreshBody)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 有効なリフレッシュトークンで更新した時、レスポンスに新しいアクセストークンが含まれること。
func TestAuthHandler_Refresh_Success_ReturnsAccessToken(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(okRefresh, validRefreshBody)

	// Assert
	assert.Contains(t, rec.Body.String(), "new-access-tok")
}

// リフレッシュトークンが無効な時、401 が返ること。
func TestAuthHandler_Refresh_Invalid(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(func(_ context.Context, _ string) (string, error) {
		return "", service.ErrInvalidCredentials
	}, validRefreshBody)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// refresh が欠けている時、サービスを呼ばず 400 が返ること。
func TestAuthHandler_Refresh_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(func(_ context.Context, _ string) (string, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return "", nil
	}, `{}`) // refresh 欠落

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// JSON として壊れたボディの時、400 が返ること。
func TestAuthHandler_Refresh_MalformedBody(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(func(_ context.Context, _ string) (string, error) {
		return "", nil
	}, `{not json`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
