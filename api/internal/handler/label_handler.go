package handler

import (
	"net/http"

	"recipe-backend/internal/dto/response"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type LabelHandler struct {
	svc service.LabelService
}

func NewLabelHandler(svc service.LabelService) *LabelHandler {
	return &LabelHandler{svc: svc}
}

// List は GET /api/label/。全ラベルを返す。
func (h *LabelHandler) List(c echo.Context) error {
	labels, err := h.svc.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	out := make([]response.LabelResponse, 0, len(labels))
	for i := range labels {
		out = append(out, response.LabelResponse{ID: labels[i].ID, Name: labels[i].Name})
	}
	return c.JSON(http.StatusOK, out)
}
