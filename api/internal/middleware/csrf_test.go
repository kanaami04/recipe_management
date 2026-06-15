package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appmw "recipe-backend/internal/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// serveWithCSRF は RequireCustomHeader を通して method/header でリクエストし、ステータスを返す。
func serveWithCSRF(method string, withHeader bool) int {
	e := echo.New()
	e.Use(appmw.RequireCustomHeader())
	e.Add(method, "/api/x", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(method, "/api/x", nil)
	if withHeader {
		req.Header.Set(appmw.CSRFHeaderName, "XMLHttpRequest")
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code
}

// 状態変更メソッドでカスタムヘッダが無い時、403 が返ること。
func TestRequireCustomHeader_PostWithoutHeader_Forbidden(t *testing.T) {
	// Arrange & Act
	code := serveWithCSRF(http.MethodPost, false)

	// Assert
	assert.Equal(t, http.StatusForbidden, code)
}

// 状態変更メソッドでカスタムヘッダがある時、通過すること。
func TestRequireCustomHeader_PostWithHeader_OK(t *testing.T) {
	// Arrange & Act
	code := serveWithCSRF(http.MethodPost, true)

	// Assert
	assert.Equal(t, http.StatusOK, code)
}

// 安全メソッド(GET)はヘッダが無くても通過すること。
func TestRequireCustomHeader_GetWithoutHeader_OK(t *testing.T) {
	// Arrange & Act
	code := serveWithCSRF(http.MethodGet, false)

	// Assert
	assert.Equal(t, http.StatusOK, code)
}

// DELETE でカスタムヘッダが無い時、403 が返ること。
func TestRequireCustomHeader_DeleteWithoutHeader_Forbidden(t *testing.T) {
	// Arrange & Act
	code := serveWithCSRF(http.MethodDelete, false)

	// Assert
	assert.Equal(t, http.StatusForbidden, code)
}
