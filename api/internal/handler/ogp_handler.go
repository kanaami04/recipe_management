package handler

import (
	"net/http"

	"recipe-backend/internal/dto/response"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type OgpHandler struct {
	svc service.OgpService
}

func NewOgpHandler(svc service.OgpService) *OgpHandler {
	return &OgpHandler{svc: svc}
}

// Fetch は GET /api/ogp/?url=...。外部レシピ URL のサムネイル画像・タイトルを返す。
func (h *OgpHandler) Fetch(c echo.Context) error {
	rawURL := c.QueryParam("url")
	if rawURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "url is required")
	}
	meta, err := h.svc.Fetch(c.Request().Context(), rawURL)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, response.OgpResponse{Image: meta.Image, Title: meta.Title})
}
