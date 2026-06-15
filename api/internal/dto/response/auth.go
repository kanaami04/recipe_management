package response

import "recipe-backend/internal/apigen"

// 構造体定義は openapi.yaml から生成する (api ADR-0005)。生成型を再エクスポートする。
type (
	TokenResponse   = apigen.TokenResponse
	RefreshResponse = apigen.RefreshResponse
)
