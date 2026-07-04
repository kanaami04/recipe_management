package handler

import (
	"net/http"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
	"recipe-backend/internal/dto/response"
	appmw "recipe-backend/internal/middleware"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	svc     service.UserService
	avatars domain.AvatarStorage
}

func NewUserHandler(svc service.UserService, avatars domain.AvatarStorage) *UserHandler {
	return &UserHandler{svc: svc, avatars: avatars}
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
	return c.JSON(http.StatusOK, response.ToUserInfo(u, h.avatars))
}

// List は GET /api/users/。共有先選択用のユーザー一覧を返す（自分自身は除く）。
func (h *UserHandler) List(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	users, err := h.svc.List(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	return c.JSON(http.StatusOK, response.ToUserList(users))
}

// Update は PUT /api/user_info/。プロフィール(ユーザー名)を更新する。
func (h *UserHandler) Update(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.UpdateUserRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.svc.UpdateProfile(c.Request().Context(), userID, req.Username)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToUserInfo(u, h.avatars))
}

// ChangeEmail は PUT /api/user_info/email/。現在のパスワード確認のうえメールを変更する。
func (h *UserHandler) ChangeEmail(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.ChangeEmailRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.svc.ChangeEmail(c.Request().Context(), userID, req.Email, req.Password)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToUserInfo(u, h.avatars))
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

// CreateAvatarUploadURL は POST /api/user_info/avatar/。
// プロフィール画像アップロード用の署名付き URL を発行する。
func (h *UserHandler) CreateAvatarUploadURL(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.CreateAvatarUploadUrlRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	uploadURL, key, err := h.svc.CreateAvatarUploadURL(c.Request().Context(), userID, string(req.ContentType))
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.AvatarUploadUrlResponse{UploadUrl: uploadURL, Key: key})
}

// ConfirmAvatar は PUT /api/user_info/avatar/。
// アップロード済みの画像をプロフィール画像として確定する。
func (h *UserHandler) ConfirmAvatar(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.ConfirmAvatarRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.svc.ConfirmAvatar(c.Request().Context(), userID, req.Key)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToUserInfo(u, h.avatars))
}

// DeleteAvatar は DELETE /api/user_info/avatar/。プロフィール画像を削除する。
func (h *UserHandler) DeleteAvatar(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	u, err := h.svc.DeleteAvatar(c.Request().Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToUserInfo(u, h.avatars))
}
