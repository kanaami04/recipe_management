package response

import (
	"recipe-backend/internal/apigen"
	"recipe-backend/internal/domain"
)

// 構造体定義は openapi.yaml から生成する。生成型を再エクスポートする。
type (
	ShoppingListResponse     = apigen.ShoppingListResponse
	ShoppingListItemResponse = apigen.ShoppingListItemResponse
)

// ToShoppingListResponse は domain.ShoppingList を API 契約の DTO へ変換する。
// avatars は owner/shared_user の avatar_url 組み立てに使う。
func ToShoppingListResponse(l *domain.ShoppingList, avatars domain.AvatarStorage) ShoppingListResponse {
	items := make([]ShoppingListItemResponse, 0, len(l.Items))
	for i := range l.Items {
		items = append(items, ShoppingListItemResponse{
			ID:      l.Items[i].ID,
			Name:    l.Items[i].Name,
			Checked: l.Items[i].Checked,
		})
	}

	shared := make([]UserListItem, 0, len(l.SharedUsers))
	for i := range l.SharedUsers {
		shared = append(shared, ToUserListItem(&l.SharedUsers[i], avatars))
	}

	return ShoppingListResponse{
		ID:         l.ID,
		Owner:      ToUserListItem(&l.Owner, avatars),
		SharedUser: shared,
		Items:      items,
	}
}
