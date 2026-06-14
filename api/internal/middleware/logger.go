package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// RequestLogger は Echo のリクエストを slog で構造化ログ出力するミドルウェア。
// echomw.Logger() の代わりに使う。
func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogStatus:    true,
		LogMethod:    true,
		LogURI:       true,
		LogLatency:   true,
		LogError:     true,
		LogRequestID: true,
		HandleError:  true, // エラーをグローバルエラーハンドラへ伝播させる
		LogValuesFunc: func(c echo.Context, v echomw.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("request_id", v.RequestID),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
			}
			ctx := c.Request().Context()
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				logger.LogAttrs(ctx, slog.LevelError, "request", attrs...)
			} else {
				logger.LogAttrs(ctx, slog.LevelInfo, "request", attrs...)
			}
			return nil
		},
	})
}
