package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appmw "recipe-backend/internal/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// serveWithOriginVerify は RequireOriginVerify(secret) を通して header 付きでリクエストし、ステータスを返す。
func serveWithOriginVerify(secret, header string) int {
	e := echo.New()
	e.Use(appmw.RequireOriginVerify(secret))
	e.GET("/api/x", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	if header != "" {
		req.Header.Set(appmw.OriginVerifyHeaderName, header)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code
}

// ヘッダが期待値と一致する時、通過すること。
func TestRequireOriginVerify_MatchingHeader_OK(t *testing.T) {
	// Arrange & Act
	code := serveWithOriginVerify("secret-value", "secret-value")

	// Assert
	assert.Equal(t, http.StatusOK, code)
}

// ヘッダが無い時、403 が返ること。
func TestRequireOriginVerify_MissingHeader_Forbidden(t *testing.T) {
	// Arrange & Act
	code := serveWithOriginVerify("secret-value", "")

	// Assert
	assert.Equal(t, http.StatusForbidden, code)
}

// ヘッダが期待値と異なる時、403 が返ること。
func TestRequireOriginVerify_WrongHeader_Forbidden(t *testing.T) {
	// Arrange & Act
	code := serveWithOriginVerify("secret-value", "wrong-value")

	// Assert
	assert.Equal(t, http.StatusForbidden, code)
}

// secret が空(未設定)の時、ヘッダが無くても通過すること(ローカル開発向け)。
func TestRequireOriginVerify_EmptySecret_OK(t *testing.T) {
	// Arrange & Act
	code := serveWithOriginVerify("", "")

	// Assert
	assert.Equal(t, http.StatusOK, code)
}
