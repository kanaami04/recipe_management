package middleware

import (
	"recipe-backend/internal/pkg/requestid"

	"github.com/labstack/echo/v4"
)

// RequestIDToContext は Echo の RequestID ミドルウェアが設定した X-Request-Id を
// リクエスト context に載せ、下位層（GORM ログ等）から参照できるようにする。
// echomw.RequestID() の後段に置くこと。
func RequestIDToContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rid := c.Response().Header().Get(echo.HeaderXRequestID)
			if rid != "" {
				ctx := requestid.WithRequestID(c.Request().Context(), rid)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	}
}
