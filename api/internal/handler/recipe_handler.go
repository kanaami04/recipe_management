package handler

import (
	"errors"
	"net/http"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
	"recipe-backend/internal/dto/response"
	appmw "recipe-backend/internal/middleware"
	"recipe-backend/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RecipeHandler struct {
	svc     service.RecipeService
	avatars domain.AvatarStorage
}

func NewRecipeHandler(svc service.RecipeService, avatars domain.AvatarStorage) *RecipeHandler {
	return &RecipeHandler{svc: svc, avatars: avatars}
}

// List は GET /api/recipes/。自分が所有 or 共有されたレシピ一覧を返す。
func (h *RecipeHandler) List(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	recipes, err := h.svc.List(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	out := make([]response.RecipeResponse, 0, len(recipes))
	for i := range recipes {
		out = append(out, response.ToRecipeResponse(&recipes[i], h.avatars))
	}
	return c.JSON(http.StatusOK, out)
}

// Create は POST /api/recipes/。
func (h *RecipeHandler) Create(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	req, err := bindRecipe(c)
	if err != nil {
		return err
	}
	recipe, err := h.svc.Create(c.Request().Context(), userID, *req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusCreated, response.ToRecipeResponse(recipe, h.avatars))
}

// Update は PUT /api/recipes/:id/。
func (h *RecipeHandler) Update(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	req, err := bindRecipe(c)
	if err != nil {
		return err
	}
	recipe, err := h.svc.Update(c.Request().Context(), userID, id, *req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToRecipeResponse(recipe, h.avatars))
}

// Delete は DELETE /api/recipes/:id/。
func (h *RecipeHandler) Delete(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), userID, id); err != nil {
		return mapServiceError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Reorder は PUT /api/recipes/reorder/。ユーザーごとの一覧表示順を更新する。
func (h *RecipeHandler) Reorder(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	var req request.ReorderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.Reorder(c.Request().Context(), userID, req.RecipeIds); err != nil {
		return mapServiceError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Archive は PUT /api/recipes/:id/archive/。ユーザーごとのアーカイブ状態を更新する。
func (h *RecipeHandler) Archive(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req request.ArchiveRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := h.svc.SetArchived(c.Request().Context(), userID, id, req.ArchiveFlg); err != nil {
		return mapServiceError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func bindRecipe(c echo.Context) (*request.RecipeRequest, error) {
	var req request.RecipeRequest
	if err := bindAndValidate(c, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// bindAndValidate はリクエストボディを req に bind し、バリデーションまで行う。
func bindAndValidate(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func parseID(c echo.Context) (string, error) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	return id, nil
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "you do not have permission to perform this action")
	case errors.Is(err, service.ErrSharedUserNotFound):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrDuplicate):
		return echo.NewHTTPError(http.StatusConflict, "already exists")
	case errors.Is(err, service.ErrUserAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrIncorrectPassword):
		return echo.NewHTTPError(http.StatusBadRequest, "現在のパスワードが違います")
	case errors.Is(err, service.ErrInvalidURL):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid url")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}
