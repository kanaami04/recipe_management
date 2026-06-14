package handler

import (
	"net/http"

	"recipe-backend/internal/dto/response"
	appmw "recipe-backend/internal/middleware"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Info は GET /api/user_info/。ログインユーザー情報を返す。
func (h *UserHandler) Info(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	u, err := h.svc.GetByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if u == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return c.JSON(http.StatusOK, response.ToUserInfo(u))
}

// List は GET /api/users/。共有先選択用に全ユーザーを返す。
func (h *UserHandler) List(c echo.Context) error {
	users, err := h.svc.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	return c.JSON(http.StatusOK, response.ToUserList(users))
}
