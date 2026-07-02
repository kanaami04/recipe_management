package handler

import (
	"net/http"

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

// List は GET /api/label/。自分が閲覧できるレシピに付いたラベル名を返す。
func (h *LabelHandler) List(c echo.Context) error {
	userID := appmw.UserIDFromContext(c)
	names, err := h.svc.List(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	out := make([]response.LabelResponse, 0, len(names))
	for _, name := range names {
		out = append(out, response.LabelResponse{Name: name})
	}
	return c.JSON(http.StatusOK, out)
}
