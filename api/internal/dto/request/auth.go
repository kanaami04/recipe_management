package request

// TokenRequest は POST /api/token/ のリクエスト。
type TokenRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest は POST /api/token/refresh/ のリクエスト。
type RefreshRequest struct {
	Refresh string `json:"refresh" validate:"required"`
}

// RegisterRequest は POST /api/auth/register/ のリクエスト。
type RegisterRequest struct {
	Username string `json:"username" validate:"required,max=50"`
	Email    string `json:"email" validate:"required,email,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}
