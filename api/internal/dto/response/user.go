package response

import (
	"recipe-backend/internal/apigen"
	"recipe-backend/internal/domain"
)

// 構造体定義は openapi.yaml から生成する。生成型を再エクスポートする。
// UserListItem は recipe の owner/shared_user でも使う。
type (
	UserInfoResponse        = apigen.UserInfoResponse
	UserListItem            = apigen.UserListItem
	AvatarUploadUrlResponse = apigen.AvatarUploadUrlResponse
)

// ToUserInfo は domain.User を API 契約の DTO へ変換する。
// avatars は avatar_url の組み立て(相対パス/絶対 URL の切り替え)に使う。
func ToUserInfo(u *domain.User, avatars domain.AvatarStorage) UserInfoResponse {
	var avatarURL *string
	if u.AvatarKey != nil {
		url := avatars.PublicURL(*u.AvatarKey)
		avatarURL = &url
	}
	return UserInfoResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.In(jst).Format(dateLayout),
		AvatarUrl: avatarURL,
	}
}

func ToUserList(users []domain.User) []UserListItem {
	out := make([]UserListItem, 0, len(users))
	for i := range users {
		out = append(out, UserListItem{ID: users[i].ID, Username: users[i].Username})
	}
	return out
}
