package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// OGP メタタグを含む HTML を渡した時、og:image と og:title を取り出せること。
func TestParseOGP_ExtractsImageAndTitle(t *testing.T) {
	// Arrange
	html := `<html><head>
		<meta property="og:title" content="絶品カレー">
		<meta property="og:image" content="https://img.example/curry.jpg">
	</head><body>本文</body></html>`

	// Act
	meta := parseOGP(strings.NewReader(html))

	// Assert
	assert.Equal(t, OgpMetadata{Image: "https://img.example/curry.jpg", Title: "絶品カレー"}, meta)
}

// property と name を両方持つ複合タグの時、property 側(og:image)を優先して拾うこと。
func TestParseOGP_PrefersPropertyOverName(t *testing.T) {
	// Arrange
	html := `<head><meta property="og:image" name="twitter:image" content="https://img.example/x.jpg"></head>`

	// Act
	meta := parseOGP(strings.NewReader(html))

	// Assert
	assert.Equal(t, "https://img.example/x.jpg", meta.Image)
}

// OGP メタタグが無い HTML を渡した時、空のメタデータを返すこと。
func TestParseOGP_NoTags_ReturnsEmpty(t *testing.T) {
	// Arrange
	html := `<html><head><title>ただのページ</title></head><body></body></html>`

	// Act
	meta := parseOGP(strings.NewReader(html))

	// Assert
	assert.Equal(t, OgpMetadata{}, meta)
}

// プライベート/ループバック等の宛先 IP は非公開と判定されること(SSRF 対策)。
func TestIsPublicIP(t *testing.T) {
	cases := map[string]struct {
		ip   string
		want bool
	}{
		"グローバル":          {"93.184.216.34", true},
		"ループバック":         {"127.0.0.1", false},
		"プライベート10系":      {"10.0.0.5", false},
		"プライベート192系":     {"192.168.1.1", false},
		"リンクローカル":        {"169.254.169.254", false},
		"CGN(100.64/10)": {"100.100.100.200", false},
		"予約(class E)":    {"240.0.0.1", false},
		"未指定":            {"0.0.0.0", false},
		"IPv6ループバック":     {"::1", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Arrange
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip)

			// Act & Assert
			assert.Equal(t, tc.want, isPublicIP(ip))
		})
	}
}

// http/https 以外のスキームの URL を渡した時、ErrInvalidURL を返すこと。
func TestFetch_InvalidScheme(t *testing.T) {
	// Arrange
	svc := NewOgpService()

	// Act
	_, err := svc.Fetch(context.Background(), "ftp://example.com/x")

	// Assert
	assert.True(t, errors.Is(err, ErrInvalidURL))
}

// ループバック宛の URL は SSRF 対策で接続が拒否され、空のメタデータを返すこと。
func TestFetch_LoopbackBlocked(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<meta property="og:image" content="https://img.example/x.jpg">`))
	}))
	defer srv.Close()
	svc := NewOgpService()

	// Act
	meta, err := svc.Fetch(context.Background(), srv.URL)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, OgpMetadata{}, *meta)
}
