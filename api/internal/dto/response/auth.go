package response

// TokenResponse は POST /api/token/ のレスポンス。
type TokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// RefreshResponse は POST /api/token/refresh/ のレスポンス。
type RefreshResponse struct {
	Access string `json:"access"`
}
