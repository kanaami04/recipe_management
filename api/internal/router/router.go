package router

import (
	"recipe-backend/internal/handler"
	appmw "recipe-backend/internal/middleware"
	jwtpkg "recipe-backend/internal/pkg/jwt"

	"github.com/labstack/echo/v4"
)

// Register は API のルートを定義する（ルーティングのみを担う）。
// 配線・ミドルウェア適用は合成ルートの app パッケージが行う。
func Register(e *echo.Echo, h *handler.Handlers, jwtManager *jwtpkg.Manager) {
	api := e.Group("/api")

	// 認証不要
	api.POST("/token/", h.Auth.Token)
	api.POST("/token/refresh/", h.Auth.Refresh)

	// 認証必須
	auth := api.Group("", appmw.JWT(jwtManager))
	auth.GET("/user_info/", h.User.Info)
	auth.GET("/users/", h.User.List)
	auth.GET("/label/", h.Label.List)
	auth.GET("/recipes/", h.Recipe.List)
	auth.POST("/recipes/", h.Recipe.Create)
	auth.PUT("/recipes/:id/", h.Recipe.Update)
	auth.DELETE("/recipes/:id/", h.Recipe.Delete)
}
