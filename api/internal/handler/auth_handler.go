package handler

import (
	"errors"
	"net/http"

	"recipe-backend/internal/dto/request"
	"recipe-backend/internal/dto/response"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
)

// refresh トークンを載せる Cookie の属性。
const (
	refreshCookieName = "refresh_token"
	// Path を refresh エンドポイントに絞り、他リクエストには送らない。
	refreshCookiePath = "/api/token/refresh/"
	// refresh TTL(7日)に合わせる(jwt パッケージの RefreshTTL と同値)。
	refreshCookieMaxAge = 7 * 24 * 60 * 60
)

type AuthHandler struct {
	svc service.AuthService
	// cookieSecure は refresh Cookie に Secure を付けるか(本番 true / dev false)。
	cookieSecure bool
}

func NewAuthHandler(svc service.AuthService, cookieSecure bool) *AuthHandler {
	return &AuthHandler{svc: svc, cookieSecure: cookieSecure}
}

// refreshCookie は refresh トークン用 Cookie を組み立てる。maxAge<0 で即時失効。
func (h *AuthHandler) refreshCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     refreshCookieName,
		Value:    value,
		Path:     refreshCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

// Token は POST /api/token/。access を body、refresh を httpOnly Cookie で返す。
func (h *AuthHandler) Token(c echo.Context) error {
	var req request.TokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	access, refresh, err := h.svc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "no active account found with the given credentials")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	c.SetCookie(h.refreshCookie(refresh, refreshCookieMaxAge))
	return c.JSON(http.StatusOK, response.TokenResponse{Access: access})
}

// Refresh は POST /api/token/refresh/。refresh を Cookie から読み、新しい access を返す。
func (h *AuthHandler) Refresh(c echo.Context) error {
	cookie, err := c.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "token is invalid or expired")
	}

	access, err := h.svc.Refresh(c.Request().Context(), cookie.Value)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "token is invalid or expired")
	}
	return c.JSON(http.StatusOK, response.RefreshResponse{Access: access})
}

// Logout は POST /api/auth/logout/。refresh Cookie を失効させる。
func (h *AuthHandler) Logout(c echo.Context) error {
	c.SetCookie(h.refreshCookie("", -1))
	return c.NoContent(http.StatusNoContent)
}

// Register は POST /api/auth/register/。新規ユーザーを作成する。
func (h *AuthHandler) Register(c echo.Context) error {
	var req request.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.svc.Register(c.Request().Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	return c.JSON(http.StatusCreated, response.ToUserInfo(user))
}
