package handler

import (
	"net/http"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
	"recipe-backend/internal/dto/response"
	appmw "recipe-backend/internal/middleware"
	"recipe-backend/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ShareGroupHandler struct {
	svc     service.ShareGroupService
	avatars domain.AvatarStorage
}

func NewShareGroupHandler(svc service.ShareGroupService, avatars domain.AvatarStorage) *ShareGroupHandler {
	return &ShareGroupHandler{svc: svc, avatars: avatars}
}

// Get は GET /api/share_group/。未所属なら 404。
func (h *ShareGroupHandler) Get(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	ctx := c.Request().Context()
	group, err := h.svc.GetMine(ctx, userID)
	if err != nil {
		return mapServiceError(err)
	}
	if group == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	sharing, err := h.svc.ShoppingListSharing(ctx, userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShareGroupResponse(group, userID, sharing, h.avatars))
}

// Create は POST /api/share_group/。
func (h *ShareGroupHandler) Create(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.CreateShareGroupRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	group, err := h.svc.Create(c.Request().Context(), userID, req.Name)
	if err != nil {
		return mapServiceError(err)
	}
	// 作成直後の所有者は常に統合設定 true(自分のリストがそのままグループの共有リストになる)。
	return c.JSON(http.StatusCreated, response.ToShareGroupResponse(group, userID, true, h.avatars))
}

// Join は POST /api/share_group/join/。
func (h *ShareGroupHandler) Join(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.JoinShareGroupRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	group, err := h.svc.Join(c.Request().Context(), userID, req.InviteCode, req.ShareShoppingList)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShareGroupResponse(group, userID, req.ShareShoppingList, h.avatars))
}

// UpdateShoppingListSharing は PUT /api/share_group/shopping_list_sharing/。
func (h *ShareGroupHandler) UpdateShoppingListSharing(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.UpdateShoppingListSharingRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.svc.SetShoppingListSharing(ctx, userID, req.ShareShoppingList); err != nil {
		return mapServiceError(err)
	}
	group, err := h.svc.GetMine(ctx, userID)
	if err != nil {
		return mapServiceError(err)
	}
	if group == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	return c.JSON(http.StatusOK, response.ToShareGroupResponse(group, userID, req.ShareShoppingList, h.avatars))
}

// Leave は POST /api/share_group/leave/。所有者が抜けると解散する。
func (h *ShareGroupHandler) Leave(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	if err := h.svc.Leave(c.Request().Context(), userID); err != nil {
		return mapServiceError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// RegenerateInviteCode は POST /api/share_group/invite_code/。所有者のみ。
func (h *ShareGroupHandler) RegenerateInviteCode(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	group, err := h.svc.RegenerateInviteCode(c.Request().Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	// 再発行できるのは所有者のみ(統合設定は常に true)。
	return c.JSON(http.StatusOK, response.ToShareGroupResponse(group, userID, true, h.avatars))
}

// RemoveMember は DELETE /api/share_group/members/:user_id/。所有者のみ。
func (h *ShareGroupHandler) RemoveMember(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	targetID := c.Param("user_id")
	if _, err := uuid.Parse(targetID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	if err := h.svc.RemoveMember(c.Request().Context(), userID, targetID); err != nil {
		return mapServiceError(err)
	}
	return c.NoContent(http.StatusNoContent)
}
