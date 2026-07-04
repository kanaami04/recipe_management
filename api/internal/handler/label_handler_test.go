package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

const testLabelID = "00000000-0000-0000-0000-000000000009"

// jsonRequest は JSON ボディ付きのリクエストを作る(ラベルは認証 userID を検証に使わない)。
func jsonRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// serveLabelList は listFn を差し替えた LabelHandler に GET /api/label/ し結果を返す。
func serveLabelList(listFn func(context.Context, string) ([]domain.Label, error)) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewLabelHandler(&mockLabelService{listFn: listFn})
	e.GET("/api/label/", h.List)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/label/", nil))
	return rec
}

// ラベル一覧を取得した時、200 が返ること。
func TestLabelHandler_List_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveLabelList(func(_ context.Context, _ string) ([]domain.Label, error) {
		return []domain.Label{{ID: "l1", Name: "和食"}}, nil
	})

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ラベル一覧を取得した時、レスポンスにラベル名が含まれること。
func TestLabelHandler_List_ReturnsLabelsInBody(t *testing.T) {
	// Arrange & Act
	rec := serveLabelList(func(_ context.Context, _ string) ([]domain.Label, error) {
		return []domain.Label{{ID: "l1", Name: "和食"}}, nil
	})

	// Assert
	assert.Contains(t, rec.Body.String(), "和食")
}

// サービスがエラーを返した時、500 が返ること。
func TestLabelHandler_List_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveLabelList(func(_ context.Context, _ string) ([]domain.Label, error) {
		return nil, assert.AnError
	})

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// serveLabelCreate は createFn を差し替えた LabelHandler に POST /api/label/ し結果を返す。
func serveLabelCreate(createFn func(context.Context, string, string) (*domain.Label, error), body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewLabelHandler(&mockLabelService{createFn: createFn})
	e.POST("/api/label/", h.Create)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/label/", body))
	return rec
}

// ラベルを作成した時、201 が返ること。
func TestLabelHandler_Create_Returns201(t *testing.T) {
	// Arrange & Act
	rec := serveLabelCreate(func(_ context.Context, _, name string) (*domain.Label, error) {
		return &domain.Label{ID: "l1", Name: name}, nil
	}, `{"name":"和食"}`)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
}

// 名前が空の時、サービスを呼ばず 400 が返ること。
func TestLabelHandler_Create_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveLabelCreate(func(_ context.Context, _, _ string) (*domain.Label, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, `{"name":""}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 同名ラベルが既にありサービスが ErrDuplicate を返した時、409 が返ること。
func TestLabelHandler_Create_Duplicate(t *testing.T) {
	// Arrange & Act
	rec := serveLabelCreate(func(_ context.Context, _, _ string) (*domain.Label, error) {
		return nil, service.ErrDuplicate
	}, `{"name":"和食"}`)

	// Assert
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// serveLabelUpdate は renameFn を差し替えた LabelHandler に PUT /api/label/:id/ し結果を返す。
func serveLabelUpdate(renameFn func(context.Context, string, string, string) (*domain.Label, error), idParam, body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewLabelHandler(&mockLabelService{renameFn: renameFn})
	e.PUT("/api/label/:id/", h.Update)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPut, "/api/label/"+idParam+"/", body))
	return rec
}

// ラベルを改名した時、200 が返ること。
func TestLabelHandler_Update_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveLabelUpdate(func(_ context.Context, _, _, name string) (*domain.Label, error) {
		return &domain.Label{ID: testLabelID, Name: name}, nil
	}, testLabelID, `{"name":"日本料理"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// id が UUID でない時、サービスを呼ばず 400 が返ること。
func TestLabelHandler_Update_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveLabelUpdate(func(_ context.Context, _, _, _ string) (*domain.Label, error) {
		t.Fatal("id 不正時に service を呼んではいけない")
		return nil, nil
	}, "not-a-uuid", `{"name":"日本料理"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 他人のラベルでサービスが ErrForbidden を返した時、403 が返ること。
func TestLabelHandler_Update_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveLabelUpdate(func(_ context.Context, _, _, _ string) (*domain.Label, error) {
		return nil, service.ErrForbidden
	}, testLabelID, `{"name":"日本料理"}`)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// 存在しないラベルでサービスが ErrNotFound を返した時、404 が返ること。
func TestLabelHandler_Update_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveLabelUpdate(func(_ context.Context, _, _, _ string) (*domain.Label, error) {
		return nil, service.ErrNotFound
	}, testLabelID, `{"name":"日本料理"}`)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// 改名先が既にありサービスが ErrDuplicate を返した時、409 が返ること。
func TestLabelHandler_Update_Duplicate(t *testing.T) {
	// Arrange & Act
	rec := serveLabelUpdate(func(_ context.Context, _, _, _ string) (*domain.Label, error) {
		return nil, service.ErrDuplicate
	}, testLabelID, `{"name":"洋食"}`)

	// Assert
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// serveLabelDelete は deleteFn を差し替えた LabelHandler に DELETE /api/label/:id/ し結果を返す。
func serveLabelDelete(deleteFn func(context.Context, string, string) error, idParam string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewLabelHandler(&mockLabelService{deleteFn: deleteFn})
	e.DELETE("/api/label/:id/", h.Delete)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodDelete, "/api/label/"+idParam+"/", ""))
	return rec
}

// ラベルを削除した時、204 が返ること。
func TestLabelHandler_Delete_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveLabelDelete(func(_ context.Context, _, _ string) error {
		return nil
	}, testLabelID)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// 他人のラベルでサービスが ErrForbidden を返した時、403 が返ること。
func TestLabelHandler_Delete_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveLabelDelete(func(_ context.Context, _, _ string) error {
		return service.ErrForbidden
	}, testLabelID)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// 存在しないラベルでサービスが ErrNotFound を返した時、404 が返ること。
func TestLabelHandler_Delete_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveLabelDelete(func(_ context.Context, _, _ string) error {
		return service.ErrNotFound
	}, testLabelID)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
