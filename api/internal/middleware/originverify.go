package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"
)

// OriginVerifyHeaderName は CloudFront がオリジンへのリクエストに付与する
// シークレットヘッダ名。
const OriginVerifyHeaderName = "X-Origin-Verify"

// RequireOriginVerify は X-Origin-Verify ヘッダが期待値と一致することを必須にする
// 。Lambda Function URL は認証なし(AuthType: NONE)で公開されるため、
// CloudFront を経由しない直接アクセスをアプリ層で遮断する。
// secret が空の場合は検証しない(ローカル開発向け)。
func RequireOriginVerify(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if secret == "" {
				return next(c)
			}
			header := c.Request().Header.Get(OriginVerifyHeaderName)
			if subtle.ConstantTimeCompare([]byte(header), []byte(secret)) != 1 {
				return echo.NewHTTPError(http.StatusForbidden, "forbidden")
			}
			return next(c)
		}
	}
}
