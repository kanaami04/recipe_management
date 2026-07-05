package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appmw "recipe-backend/internal/middleware"
	jwtpkg "recipe-backend/internal/pkg/jwt"
	"recipe-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

type mockOgpService struct {
	fetchFn func(ctx context.Context, rawURL string) (*service.OgpMetadata, error)
}

func (m *mockOgpService) Fetch(ctx context.Context, rawURL string) (*service.OgpMetadata, error) {
	return m.fetchFn(ctx, rawURL)
}

// serveOgp は fetchFn を差し替えた OgpHandler に、認証付きで GET /api/ogp/ し結果を返す。
func serveOgp(t *testing.T, fetchFn func(context.Context, string) (*service.OgpMetadata, error), query string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewOgpHandler(&mockOgpService{fetchFn: fetchFn})
	e.GET("/api/ogp/", h.Fetch, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodGet, "/api/ogp/"+query, ""))
	return rec
}

// url パラメータが無い時、サービスを呼ばず 400 を返すこと。
func TestOgpHandler_MissingURL_Returns400(t *testing.T) {
	// Arrange & Act
	rec := serveOgp(t, func(context.Context, string) (*service.OgpMetadata, error) {
		t.Fatal("url が空のときはサービスを呼んではいけない")
		return nil, nil
	}, "")

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// サービスが ErrInvalidURL を返す時、400 を返すこと。
func TestOgpHandler_InvalidURL_Returns400(t *testing.T) {
	// Arrange & Act
	rec := serveOgp(t, func(context.Context, string) (*service.OgpMetadata, error) {
		return nil, service.ErrInvalidURL
	}, "?url=x")

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 取得成功時、200 を返すこと。
func TestOgpHandler_Success_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveOgp(t, func(context.Context, string) (*service.OgpMetadata, error) {
		return &service.OgpMetadata{Image: "https://img.example/x.jpg", Title: "カレー"}, nil
	}, "?url=x")

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}
