package request

import "recipe-backend/internal/apigen"

// 構造体定義は openapi.yaml から生成する。生成型を再エクスポートする。
type (
	RecipeRequest   = apigen.RecipeRequest
	LabelInput      = apigen.LabelInput
	SharedUserInput = apigen.SharedUserInput
	NameInput       = apigen.NameInput
	CookingInput    = apigen.CookingInput
	SeasonInput     = apigen.SeasonInput
)
