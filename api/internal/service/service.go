package service

import (
	"recipe-backend/internal/domain"
	jwtpkg "recipe-backend/internal/pkg/jwt"
)

// Services は全サービスを束ねる。
type Services struct {
	Auth   AuthService
	User   UserService
	Label  LabelService
	Recipe RecipeService
}

// New はリポジトリ（domain interface）から全サービスを生成する（サービス層の合成）。
func New(
	userRepo domain.UserRepository,
	labelRepo domain.LabelRepository,
	recipeRepo domain.RecipeRepository,
	jwt *jwtpkg.Manager,
) *Services {
	return &Services{
		Auth:   NewAuthService(userRepo, jwt),
		User:   NewUserService(userRepo),
		Label:  NewLabelService(labelRepo),
		Recipe: NewRecipeService(recipeRepo, userRepo),
	}
}
