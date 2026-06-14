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
