package handler

import (
	"errors"
	"net/http"

	"recipe-backend/internal/dto/request"
	"recipe-backend/internal/dto/response"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Token は POST /api/token/。username/password を検証して access/refresh を返す。
func (h *AuthHandler) Token(c echo.Context) error {
	var req request.TokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	access, refresh, err := h.svc.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "no active account found with the given credentials")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	return c.JSON(http.StatusOK, response.TokenResponse{Access: access, Refresh: refresh})
}

// Refresh は POST /api/token/refresh/。refresh から新しい access を返す。
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req request.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	access, err := h.svc.Refresh(c.Request().Context(), req.Refresh)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "token is invalid or expired")
	}
	return c.JSON(http.StatusOK, response.RefreshResponse{Access: access})
}
