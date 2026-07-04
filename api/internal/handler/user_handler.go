package handler

import (
	"net/http"

	"recipe-backend/internal/dto/request"
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

// Update は PUT /api/user_info/。プロフィール(ユーザー名・メール)を更新する。
func (h *UserHandler) Update(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.UpdateUserRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.svc.UpdateProfile(c.Request().Context(), userID, req.Username, req.Email)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToUserInfo(u))
}

// ChangePassword は PUT /api/user_info/password/。パスワードを変更する。
func (h *UserHandler) ChangePassword(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.ChangePasswordRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.ChangePassword(c.Request().Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		return mapServiceError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Delete は DELETE /api/user_info/。アカウントを削除する。
func (h *UserHandler) Delete(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	if err := h.svc.DeleteAccount(c.Request().Context(), userID); err != nil {
		return mapServiceError(err)
	}
	return c.NoContent(http.StatusNoContent)
}
