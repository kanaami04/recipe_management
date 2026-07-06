package request

import "recipe-backend/internal/apigen"

// 構造体定義は openapi.yaml から生成する。生成型を再エクスポートする。
type (
	CreateShareGroupRequest          = apigen.CreateShareGroupRequest
	JoinShareGroupRequest            = apigen.JoinShareGroupRequest
	UpdateShoppingListSharingRequest = apigen.UpdateShoppingListSharingRequest
)
