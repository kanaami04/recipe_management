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

type ShoppingListHandler struct {
	svc     service.ShoppingListService
	avatars domain.AvatarStorage
}

func NewShoppingListHandler(svc service.ShoppingListService, avatars domain.AvatarStorage) *ShoppingListHandler {
	return &ShoppingListHandler{svc: svc, avatars: avatars}
}

// Get は GET /api/shopping_list/。自分の買い物リストを取得(無ければ作成)して返す。
func (h *ShoppingListHandler) Get(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	list, err := h.svc.Get(c.Request().Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShoppingListResponse(list, h.avatars))
}

// AddItem は POST /api/shopping_list/:id/items/。項目を追加する。
func (h *ShoppingListHandler) AddItem(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req request.ShoppingListItemInput
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	list, err := h.svc.AddItem(c.Request().Context(), userID, id, req.Name)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShoppingListResponse(list, h.avatars))
}

// UpdateItem は PUT /api/shopping_list/:id/items/:item_id/。チェック状態を更新する。
func (h *ShoppingListHandler) UpdateItem(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	itemID, err := parseItemID(c)
	if err != nil {
		return err
	}
	var req request.ShoppingListItemUpdateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	list, err := h.svc.SetItemChecked(c.Request().Context(), userID, id, itemID, req.Checked)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShoppingListResponse(list, h.avatars))
}

// DeleteItem は DELETE /api/shopping_list/:id/items/:item_id/。項目を削除する。
func (h *ShoppingListHandler) DeleteItem(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	itemID, err := parseItemID(c)
	if err != nil {
		return err
	}
	list, err := h.svc.DeleteItem(c.Request().Context(), userID, id, itemID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShoppingListResponse(list, h.avatars))
}

// ClearChecked は DELETE /api/shopping_list/:id/items/checked/。チェック済みをまとめて削除する。
func (h *ShoppingListHandler) ClearChecked(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	list, err := h.svc.ClearChecked(c.Request().Context(), userID, id)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShoppingListResponse(list, h.avatars))
}

// Reorder は PUT /api/shopping_list/:id/items/reorder/。項目の表示順を更新する。
func (h *ShoppingListHandler) Reorder(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req request.ShoppingListReorderRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	list, err := h.svc.Reorder(c.Request().Context(), userID, id, req.ItemIds)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShoppingListResponse(list, h.avatars))
}

// UpdateShares は PUT /api/shopping_list/:id/shares/。共有相手を更新する。
func (h *ShoppingListHandler) UpdateShares(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req request.ShoppingListSharesRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	list, err := h.svc.UpdateShares(c.Request().Context(), userID, id, req.SharedUser)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToShoppingListResponse(list, h.avatars))
}

func parseItemID(c echo.Context) (string, error) {
	itemID := c.Param("item_id")
	if _, err := uuid.Parse(itemID); err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid item id")
	}
	return itemID, nil
}
