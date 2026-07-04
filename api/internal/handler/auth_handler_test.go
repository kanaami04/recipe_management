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
	h := NewAuthHandler(&mockAuthService{loginFn: loginFn}, false, mockAvatarStorage{})
	e.POST("/api/token/", h.Token)
	req := httptest.NewRequest(http.MethodPost, "/api/token/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// serveRegister は registerFn を差し替えた AuthHandler に POST /api/auth/register/ し、結果を返す。
func serveRegister(registerFn func(context.Context, string, string, string) (*domain.User, error), body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{registerFn: registerFn}, false, mockAvatarStorage{})
	e.POST("/api/auth/register/", h.Register)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// serveRefresh は refreshFn を差し替えた AuthHandler に POST /api/token/refresh/ する。
// cookieValue が空でなければ refresh Cookie を付ける(空なら Cookie なし)。
func serveRefresh(refreshFn func(context.Context, string) (string, error), cookieValue string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{refreshFn: refreshFn}, false, mockAvatarStorage{})
	e.POST("/api/token/refresh/", h.Refresh)
	req := httptest.NewRequest(http.MethodPost, "/api/token/refresh/", nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: cookieValue})
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// serveLogout は AuthHandler に POST /api/auth/logout/ する。
func serveLogout() *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewAuthHandler(&mockAuthService{}, false, mockAvatarStorage{})
	e.POST("/api/auth/logout/", h.Logout)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// findCookie はレスポンスから指定名の Set-Cookie を探す。
func findCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// okLogin は常に固定トークンを返す loginFn。
func okLogin(_ context.Context, _, _ string) (string, string, error) {
	return "access-tok", "refresh-tok", nil
}

// okRegister は受け取った username/email でユーザーを返す registerFn。
func okRegister(_ context.Context, u, em, _ string) (*domain.User, error) {
	return &domain.User{ID: "u1", Username: u, Email: em}, nil
}

// okRefresh は固定のアクセストークンを返す refreshFn。
func okRefresh(_ context.Context, _ string) (string, error) {
	return "new-access-tok", nil
}

const (
	validLoginBody    = `{"email":"alice@example.com","password":"pw"}`
	validRegisterBody = `{"username":"alice","email":"alice@example.com","password":"password123"}`
)

// 正しい資格情報でトークン取得した時、200 が返ること。
func TestAuthHandler_Token_Success_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveToken(okLogin, validLoginBody)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 正しい資格情報でトークン取得した時、レスポンス body にアクセストークンが含まれること。
func TestAuthHandler_Token_Success_ReturnsAccessToken(t *testing.T) {
	// Arrange & Act
	rec := serveToken(okLogin, validLoginBody)

	// Assert
	assert.Contains(t, rec.Body.String(), "access-tok")
}

// 正しい資格情報でトークン取得した時、refresh が httpOnly Cookie でセットされること。
func TestAuthHandler_Token_Success_SetsRefreshCookie(t *testing.T) {
	// Arrange & Act
	rec := serveToken(okLogin, validLoginBody)

	// Assert
	cookie := findCookie(rec, refreshCookieName)
	assert.NotNil(t, cookie)
	assert.Equal(t, "refresh-tok", cookie.Value)
	assert.True(t, cookie.HttpOnly)
}

// 正しい資格情報でトークン取得した時、refresh が body には含まれないこと。
func TestAuthHandler_Token_Success_DoesNotReturnRefreshInBody(t *testing.T) {
	// Arrange & Act
	rec := serveToken(okLogin, validLoginBody)

	// Assert
	assert.NotContains(t, rec.Body.String(), "refresh-tok")
}

// 認証情報が誤っている時、401 が返ること。
func TestAuthHandler_Token_InvalidCredentials(t *testing.T) {
	// Arrange & Act
	rec := serveToken(func(_ context.Context, _, _ string) (string, string, error) {
		return "", "", service.ErrInvalidCredentials
	}, `{"email":"alice@example.com","password":"bad"}`)

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
	rec := serveRegister(func(_ context.Context, _, _, _ string) (*domain.User, error) {
		return nil, service.ErrUserAlreadyExists
	}, validRegisterBody)

	// Assert
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// email 不正やパスワード不足の時、サービスを呼ばず 400 が返ること。
func TestAuthHandler_Register_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveRegister(func(_ context.Context, _, _, _ string) (*domain.User, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, `{"username":"alice","email":"not-an-email","password":"short"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 有効な refresh Cookie で更新した時、200 が返ること。
func TestAuthHandler_Refresh_Success_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(okRefresh, "some-refresh-token")

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 有効な refresh Cookie で更新した時、レスポンスに新しいアクセストークンが含まれること。
func TestAuthHandler_Refresh_Success_ReturnsAccessToken(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(okRefresh, "some-refresh-token")

	// Assert
	assert.Contains(t, rec.Body.String(), "new-access-tok")
}

// refresh Cookie が無効な時、401 が返ること。
func TestAuthHandler_Refresh_Invalid(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(func(_ context.Context, _ string) (string, error) {
		return "", service.ErrInvalidCredentials
	}, "bad-refresh-token")

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// refresh Cookie が無い時、サービスを呼ばず 401 が返ること。
func TestAuthHandler_Refresh_MissingCookie(t *testing.T) {
	// Arrange & Act
	rec := serveRefresh(func(_ context.Context, _ string) (string, error) {
		t.Fatal("Cookie が無い時に service を呼んではいけない")
		return "", nil
	}, "")

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ログアウトした時、204 が返ること。
func TestAuthHandler_Logout_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveLogout()

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// ログアウトした時、refresh Cookie が失効(値が空)で上書きされること。
func TestAuthHandler_Logout_ClearsRefreshCookie(t *testing.T) {
	// Arrange & Act
	rec := serveLogout()

	// Assert
	cookie := findCookie(rec, refreshCookieName)
	assert.NotNil(t, cookie)
	assert.Empty(t, cookie.Value)
}
