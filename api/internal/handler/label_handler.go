package handler

import (
	"net/http"

	"recipe-backend/internal/dto/request"
	"recipe-backend/internal/dto/response"
	appmw "recipe-backend/internal/middleware"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type LabelHandler struct {
	svc service.LabelService
}

func NewLabelHandler(svc service.LabelService) *LabelHandler {
	return &LabelHandler{svc: svc}
}

// List は GET /api/label/。自分が管理するラベル一覧を返す。
func (h *LabelHandler) List(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	labels, err := h.svc.List(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	out := make([]response.LabelItem, 0, len(labels))
	for i := range labels {
		out = append(out, response.ToLabelItem(&labels[i]))
	}
	return c.JSON(http.StatusOK, out)
}

// Create は POST /api/label/。
func (h *LabelHandler) Create(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	req, err := bindLabel(c)
	if err != nil {
		return err
	}
	label, err := h.svc.Create(c.Request().Context(), userID, req.Name)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusCreated, response.ToLabelItem(label))
}

// Update は PUT /api/label/:id/。ラベルを改名する。
func (h *LabelHandler) Update(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	id, err := parseID(c)
	if err != nil {
		return err
	}
	req, err := bindLabel(c)
	if err != nil {
		return err
	}
	label, err := h.svc.Rename(c.Request().Context(), userID, id, req.Name)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.ToLabelItem(label))
}

// Delete は DELETE /api/label/:id/。
func (h *LabelHandler) Delete(c echo.Context) error {
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

func bindLabel(c echo.Context) (*request.LabelInput, error) {
	var req request.LabelInput
	if err := c.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return &req, nil
}
