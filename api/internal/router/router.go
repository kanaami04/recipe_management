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
	{
		// 認証
		{
			api.POST("/token/", h.Auth.Token)
			api.POST("/token/refresh/", h.Auth.Refresh)
			api.POST("/auth/register/", h.Auth.Register)
			api.POST("/auth/logout/", h.Auth.Logout)
		}
	}

	// 認証必須
	{
		authorized := api.Group("", appmw.JWT(jwtManager))

		// ユーザー
		{
			authorized.GET("/user_info/", h.User.Info)
			authorized.GET("/users/", h.User.List)
		}

		// ラベル
		{
			authorized.GET("/label/", h.Label.List)
		}

		// レシピ
		{
			authorized.GET("/recipes/", h.Recipe.List)
			authorized.POST("/recipes/", h.Recipe.Create)
			authorized.PUT("/recipes/:id/", h.Recipe.Update)
			authorized.DELETE("/recipes/:id/", h.Recipe.Delete)
		}
	}
}
