package handler

import "recipe-backend/internal/service"

// Handlers は全ハンドラを束ねる。
type Handlers struct {
	Auth   *AuthHandler
	User   *UserHandler
	Label  *LabelHandler
	Recipe *RecipeHandler
}

// New は全サービスから全ハンドラを生成する（ハンドラ層の合成）。
// cookieSecure は refresh Cookie の Secure 属性に渡す。
func New(s *service.Services, cookieSecure bool) *Handlers {
	return &Handlers{
		Auth:   NewAuthHandler(s.Auth, cookieSecure),
		User:   NewUserHandler(s.User),
		Label:  NewLabelHandler(s.Label),
		Recipe: NewRecipeHandler(s.Recipe),
	}
}
