package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// OgpMetadata は外部 URL の OGP から取り出したサムネイル情報。取れない項目は空文字。
type OgpMetadata struct {
	Image string
	Title string
}

// OgpService は外部レシピ URL の OGP メタデータ(サムネイル画像・タイトル)を取得する。
// このアプリで唯一、任意の外部 URL へ HTTP する経路のため SSRF 対策を内包する。
type OgpService interface {
	Fetch(ctx context.Context, rawURL string) (*OgpMetadata, error)
}

const (
	ogpTimeout     = 5 * time.Second
	ogpMaxBodySize = 1 << 20 // 1MiB。<head> の meta を読むには十分。
	ogpUserAgent   = "recipe-management-bot/1.0 (+ogp)"
)

type ogpService struct {
	client *http.Client
}

// NewOgpService は SSRF 対策済みの HTTP クライアントを持つ OgpService を生成する。
func NewOgpService() OgpService {
	// 解決後の宛先 IP を Control で検査し、プライベート/ループバック等を拒否する。
	// Control は DNS 解決後・接続前に実際に繋ぐ IP で呼ばれるため、DNS リバインディングも防げる。
	// リダイレクト先も同じ Transport で再度 dial されるため同様に検査される。
	dialer := &net.Dialer{
		Timeout: ogpTimeout,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ip == nil || !isPublicIP(ip) {
				return fmt.Errorf("blocked non-public address: %s", address)
			}
			return nil
		},
	}
	return &ogpService{
		client: &http.Client{
			Timeout:   ogpTimeout,
			Transport: &http.Transport{DialContext: dialer.DialContext},
		},
	}
}

func (s *ogpService) Fetch(ctx context.Context, rawURL string) (*OgpMetadata, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, ErrInvalidURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, ErrInvalidURL
	}
	req.Header.Set("User-Agent", ogpUserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		// 取得失敗(到達不可・SSRF ブロック等)はサムネ無しとして扱う(best-effort)。
		return &OgpMetadata{}, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &OgpMetadata{}, nil
	}

	// 文字コードは Content-Type から判定して UTF-8 へ変換する(Shift_JIS 等の og:title 対策)。
	limited := io.LimitReader(resp.Body, ogpMaxBodySize)
	decoded, err := charset.NewReader(limited, resp.Header.Get("Content-Type"))
	if err != nil {
		decoded = limited
	}
	meta := parseOGP(decoded)
	// og:image が相対 URL のこともあるため、最終取得先を基準に絶対 URL へ解決する。
	if meta.Image != "" {
		base := resp.Request.URL
		if ref, err := base.Parse(meta.Image); err == nil {
			meta.Image = ref.String()
		}
	}
	return &meta, nil
}

// parseOGP は <head> 内の og:image / og:title を拾う。<body> に入ったら打ち切る。
func parseOGP(r io.Reader) OgpMetadata {
	var meta OgpMetadata
	z := html.NewTokenizer(r)
	for {
		switch z.Next() {
		case html.ErrorToken:
			return meta
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			switch string(name) {
			case "body":
				return meta // head を抜けたら以降に OGP は無い。
			case "meta":
				if !hasAttr {
					continue
				}
				prop, content := metaAttrs(z)
				switch prop {
				case "og:image":
					if meta.Image == "" {
						meta.Image = content
					}
				case "og:title":
					if meta.Title == "" {
						meta.Title = content
					}
				}
				if meta.Image != "" && meta.Title != "" {
					return meta
				}
			}
		}
	}
}

// metaAttrs は meta タグから OGP 用の識別子と content を取り出す。
// property と name は別々に拾い、OGP 標準の property を優先する
// (両方を持つ複合タグで og:image/og:title を取りこぼさないため)。
func metaAttrs(z *html.Tokenizer) (key, content string) {
	var property, name string
	for {
		k, v, more := z.TagAttr()
		switch string(k) {
		case "property":
			property = string(v)
		case "name":
			name = string(v)
		case "content":
			content = string(v)
		}
		if !more {
			break
		}
	}
	if property != "" {
		return property, content
	}
	return name, content
}

// reservedCIDRs は SSRF 対策で拒否する非グローバルな予約 IP 帯。
// stdlib の IsPrivate/IsLoopback 等でカバーされない CGN・予約帯を明示する。
var reservedCIDRs = func() []*net.IPNet {
	blocks := []string{
		"0.0.0.0/8",       // 現在のネットワーク
		"100.64.0.0/10",   // CGN (RFC6598)
		"192.0.0.0/24",    // IETF プロトコル割当
		"192.0.2.0/24",    // TEST-NET-1
		"198.18.0.0/15",   // ベンチマーク
		"198.51.100.0/24", // TEST-NET-2
		"203.0.113.0/24",  // TEST-NET-3
		"240.0.0.0/4",     // 予約(class E)
		"::/128",          // 未指定
		"64:ff9b:1::/48",  // NAT64
		"2001:db8::/32",   // ドキュメント用
	}
	nets := make([]*net.IPNet, 0, len(blocks))
	for _, b := range blocks {
		if _, n, err := net.ParseCIDR(b); err == nil {
			nets = append(nets, n)
		}
	}
	return nets
}()

// isPublicIP はグローバルに到達可能な IP のみ true を返す(SSRF 対策)。
// グローバルユニキャストでない(ループバック・リンクローカル・マルチキャスト等)、
// プライベート、予約帯(CGN 含む)はすべて拒否する。
func isPublicIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return false
	}
	for _, n := range reservedCIDRs {
		if n.Contains(ip) {
			return false
		}
	}
	return true
}
