package request

import "recipe-backend/internal/apigen"

// 構造体定義は openapi.yaml から生成する。生成型を再エクスポートする。
type (
	ShoppingListItemInput         = apigen.ShoppingListItemInput
	ShoppingListItemUpdateRequest = apigen.ShoppingListItemUpdateRequest
	ShoppingListReorderRequest    = apigen.ShoppingListReorderRequest
	ShoppingListBulkAddRequest    = apigen.ShoppingListBulkAddRequest
	ShoppingListBulkAddItem       = apigen.ShoppingListBulkAddItem
)
