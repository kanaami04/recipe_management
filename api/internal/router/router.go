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

		// 買い物リスト
		{
			authorized.GET("/shopping_list/", h.ShoppingList.Get)
			authorized.POST("/shopping_list/:id/items/", h.ShoppingList.AddItem)
			// checked / reorder は静的パス。:item_id より先に登録して曖昧さを避ける。
			authorized.DELETE("/shopping_list/:id/items/checked/", h.ShoppingList.ClearChecked)
			authorized.PUT("/shopping_list/:id/items/reorder/", h.ShoppingList.Reorder)
			authorized.PUT("/shopping_list/:id/items/:item_id/", h.ShoppingList.UpdateItem)
			authorized.DELETE("/shopping_list/:id/items/:item_id/", h.ShoppingList.DeleteItem)
		}

		// シェアグループ
		{
			authorized.GET("/share_group/", h.ShareGroup.Get)
			authorized.POST("/share_group/", h.ShareGroup.Create)
			authorized.POST("/share_group/join/", h.ShareGroup.Join)
			authorized.POST("/share_group/leave/", h.ShareGroup.Leave)
			authorized.POST("/share_group/invite_code/", h.ShareGroup.RegenerateInviteCode)
			authorized.PUT("/share_group/shopping_list_sharing/", h.ShareGroup.UpdateShoppingListSharing)
			authorized.DELETE("/share_group/members/:user_id/", h.ShareGroup.RemoveMember)
		}

		// OGP(外部レシピ URL のサムネイル取得)
		{
			authorized.GET("/ogp/", h.Ogp.Fetch)
		}
	}
}
