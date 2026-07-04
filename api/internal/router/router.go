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
	// CSRF 対策: 状態変更系にカスタムヘッダを必須化する。
	api := e.Group("/api", appmw.RequireCustomHeader())

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
			authorized.PUT("/user_info/", h.User.Update)
			authorized.PUT("/user_info/email/", h.User.ChangeEmail)
			authorized.PUT("/user_info/password/", h.User.ChangePassword)
			authorized.DELETE("/user_info/", h.User.Delete)
			authorized.POST("/user_info/avatar/", h.User.CreateAvatarUploadURL)
			authorized.PUT("/user_info/avatar/", h.User.ConfirmAvatar)
			authorized.DELETE("/user_info/avatar/", h.User.DeleteAvatar)
			authorized.GET("/users/", h.User.List)
		}

		// ラベル
		{
			authorized.GET("/label/", h.Label.List)
			authorized.POST("/label/", h.Label.Create)
			authorized.PUT("/label/:id/", h.Label.Update)
			authorized.DELETE("/label/:id/", h.Label.Delete)
		}

		// レシピ
		{
			authorized.GET("/recipes/", h.Recipe.List)
			authorized.POST("/recipes/", h.Recipe.Create)
			// reorder は静的パス。:id より先に登録して曖昧さを避ける。
			authorized.PUT("/recipes/reorder/", h.Recipe.Reorder)
			authorized.PUT("/recipes/:id/archive/", h.Recipe.Archive)
			authorized.PUT("/recipes/:id/", h.Recipe.Update)
			authorized.DELETE("/recipes/:id/", h.Recipe.Delete)
		}
	}
}
