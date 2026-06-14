package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
)

// serveLabelList は listFn を差し替えた LabelHandler に GET /api/label/ し結果を返す。
func serveLabelList(listFn func(context.Context) ([]domain.RecipeLabel, error)) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewLabelHandler(&mockLabelService{listFn: listFn})
	e.GET("/api/label/", h.List)
	req := httptest.NewRequest(http.MethodGet, "/api/label/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ラベル一覧を取得した時、200 が返ること。
func TestLabelHandler_List_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveLabelList(func(_ context.Context) ([]domain.RecipeLabel, error) {
		return []domain.RecipeLabel{factory.NewRecipeLabel("和食")}, nil
	})

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ラベル一覧を取得した時、レスポンスにラベル名が含まれること。
func TestLabelHandler_List_ReturnsLabelsInBody(t *testing.T) {
	// Arrange & Act
	rec := serveLabelList(func(_ context.Context) ([]domain.RecipeLabel, error) {
		return []domain.RecipeLabel{factory.NewRecipeLabel("和食")}, nil
	})

	// Assert
	assert.Contains(t, rec.Body.String(), "和食")
}

// サービスがエラーを返した時、500 が返ること。
func TestLabelHandler_List_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveLabelList(func(_ context.Context) ([]domain.RecipeLabel, error) {
		return nil, assert.AnError
	})

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
