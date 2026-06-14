package response

import "recipe-backend/internal/domain"

// UserInfoResponse は GET /api/user_info/ のレスポンス。
type UserInfoResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserListItem は GET /api/users/ の要素、および recipe の owner/shared_user で使う。
type UserListItem struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func ToUserInfo(u *domain.ApplicationUser) UserInfoResponse {
	return UserInfoResponse{ID: u.ID, Username: u.Username, Email: u.Email}
}

func ToUserList(users []domain.ApplicationUser) []UserListItem {
	out := make([]UserListItem, 0, len(users))
	for i := range users {
		out = append(out, UserListItem{ID: users[i].ID, Username: users[i].Username})
	}
	return out
}
