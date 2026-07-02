package response

import (
	"recipe-backend/internal/apigen"
	"recipe-backend/internal/domain"
)

// 構造体定義は openapi.yaml から生成する (api ADR-0005)。生成型を再エクスポートする。
// UserListItem は recipe の owner/shared_user でも使う。
type (
	UserInfoResponse = apigen.UserInfoResponse
	UserListItem     = apigen.UserListItem
)

func ToUserInfo(u *domain.User) UserInfoResponse {
	return UserInfoResponse{ID: u.ID, Username: u.Username, Email: u.Email}
}

func ToUserList(users []domain.User) []UserListItem {
	out := make([]UserListItem, 0, len(users))
	for i := range users {
		out = append(out, UserListItem{ID: users[i].ID, Username: users[i].Username})
	}
	return out
}
