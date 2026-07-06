package response

import (
	"recipe-backend/internal/apigen"
	"recipe-backend/internal/domain"
)

// 構造体定義は openapi.yaml から生成する。生成型を再エクスポートする。
type ShareGroupResponse = apigen.ShareGroupResponse

// ToShareGroupResponse は domain.ShareGroup を API 契約の DTO へ変換する。
// viewerID は is_owner の判定に、viewerShareShoppingList は viewerID 自身の買い物リスト
// 統合設定に、avatars はメンバーの avatar_url 組み立てに使う。
func ToShareGroupResponse(g *domain.ShareGroup, viewerID string, viewerShareShoppingList bool, avatars domain.AvatarStorage) ShareGroupResponse {
	members := make([]UserListItem, 0, len(g.Members))
	for i := range g.Members {
		members = append(members, ToUserListItem(&g.Members[i], avatars))
	}
	return ShareGroupResponse{
		ID:                  g.ID,
		Name:                g.Name,
		Owner:               ToUserListItem(&g.Owner, avatars),
		Members:             members,
		InviteCode:          g.InviteCode,
		InviteCodeExpiresAt: g.InviteCodeExpiresAt.In(jst).Format(dateLayout),
		IsOwner:             g.OwnerID == viewerID,
		ShareShoppingList:   viewerShareShoppingList,
	}
}
