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
func New(s *service.Services) *Handlers {
	return &Handlers{
		Auth:   NewAuthHandler(s.Auth),
		User:   NewUserHandler(s.User),
		Label:  NewLabelHandler(s.Label),
		Recipe: NewRecipeHandler(s.Recipe),
	}
}
