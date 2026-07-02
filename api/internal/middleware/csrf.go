package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// CSRFHeaderName は CSRF 対策で必須にするカスタムヘッダ名。
const CSRFHeaderName = "X-Requested-With"

// RequireCustomHeader は状態変更メソッド(POST/PUT/PATCH/DELETE)に対し、
// カスタムヘッダの存在を必須にする。
// クロスサイトのフォーム送信はカスタムヘッダを付けられないため、
// 同一オリジン + SameSite=Lax Cookie と併せて CSRF を防ぐ。
func RequireCustomHeader() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			switch c.Request().Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				if c.Request().Header.Get(CSRFHeaderName) == "" {
					return echo.NewHTTPError(http.StatusForbidden, "missing required header")
				}
			}
			return next(c)
		}
	}
}
