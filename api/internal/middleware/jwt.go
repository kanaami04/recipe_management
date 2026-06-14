package middleware

import (
	"net/http"
	"strings"

	jwtpkg "recipe-backend/internal/pkg/jwt"

	"github.com/labstack/echo/v4"
)

const userIDKey = "userID"

// JWT は Authorization: Bearer <access> を検証し、user_id を context へ格納する。
func JWT(manager *jwtpkg.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get(echo.HeaderAuthorization)
			if !strings.HasPrefix(auth, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication credentials were not provided")
			}
			tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			userID, err := manager.Parse(tokenStr, jwtpkg.TypeAccess)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "token is invalid or expired")
			}
			c.Set(userIDKey, userID)
			return next(c)
		}
	}
}

// UserIDFromContext は JWT ミドルウェアが格納した user_id を取り出す。
func UserIDFromContext(c echo.Context) uint {
	if v, ok := c.Get(userIDKey).(uint); ok {
		return v
	}
	return 0
}
