package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

const (
	testListID = "00000000-0000-0000-0000-000000000021"
	testItemID = "00000000-0000-0000-0000-000000000022"
)

func newShoppingListHandler(svc *mockShoppingListService) *ShoppingListHandler {
	return NewShoppingListHandler(svc, mockAvatarStorage{})
}

// serveShoppingListGet は getFn を差し替えたハンドラに GET /api/shopping_list/ し結果を返す。
func serveShoppingListGet(getFn func(context.Context, string) (*domain.ShoppingList, error)) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShoppingListHandler(&mockShoppingListService{getFn: getFn})
	e.GET("/api/shopping_list/", h.Get)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/shopping_list/", nil))
	return rec
}

// 買い物リストを取得した時、200 が返ること。
func TestShoppingListHandler_Get_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveShoppingListGet(func(_ context.Context, _ string) (*domain.ShoppingList, error) {
		return &domain.ShoppingList{ID: testListID}, nil
	})

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 取得した時、レスポンスに項目名が含まれること。
func TestShoppingListHandler_Get_ReturnsItemsInBody(t *testing.T) {
	// Arrange & Act
	rec := serveShoppingListGet(func(_ context.Context, _ string) (*domain.ShoppingList, error) {
		return &domain.ShoppingList{ID: testListID, Items: []domain.ShoppingListItem{{ID: testItemID, Name: "牛乳"}}}, nil
	})

	// Assert
	assert.Contains(t, rec.Body.String(), "牛乳")
}

// サービスがエラーを返した時、500 が返ること。
func TestShoppingListHandler_Get_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveShoppingListGet(func(_ context.Context, _ string) (*domain.ShoppingList, error) {
		return nil, assert.AnError
	})

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// serveAddItem は addItemFn を差し替えたハンドラに POST /api/shopping_list/:id/items/ し結果を返す。
func serveAddItem(addItemFn func(context.Context, string, string, string) (*domain.ShoppingList, error), idParam, body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShoppingListHandler(&mockShoppingListService{addItemFn: addItemFn})
	e.POST("/api/shopping_list/:id/items/", h.AddItem)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/shopping_list/"+idParam+"/items/", body))
	return rec
}

// 項目を追加した時、200 が返ること。
func TestShoppingListHandler_AddItem_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveAddItem(func(_ context.Context, _, _, name string) (*domain.ShoppingList, error) {
		return &domain.ShoppingList{ID: testListID, Items: []domain.ShoppingListItem{{ID: testItemID, Name: name}}}, nil
	}, testListID, `{"name":"牛乳"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 名前が空の時、サービスを呼ばず 400 が返ること。
func TestShoppingListHandler_AddItem_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveAddItem(func(_ context.Context, _, _, _ string) (*domain.ShoppingList, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, testListID, `{"name":""}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// id が UUID でない時、サービスを呼ばず 400 が返ること。
func TestShoppingListHandler_AddItem_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveAddItem(func(_ context.Context, _, _, _ string) (*domain.ShoppingList, error) {
		t.Fatal("id 不正時に service を呼んではいけない")
		return nil, nil
	}, "not-a-uuid", `{"name":"牛乳"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 他人のリストでサービスが ErrForbidden を返した時、403 が返ること。
func TestShoppingListHandler_AddItem_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveAddItem(func(_ context.Context, _, _, _ string) (*domain.ShoppingList, error) {
		return nil, service.ErrForbidden
	}, testListID, `{"name":"牛乳"}`)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// serveAddItems は addItemsFn を差し替えたハンドラに POST /api/shopping_list/:id/items/bulk/ し結果を返す。
func serveAddItems(addItemsFn func(context.Context, string, string, []service.NewItem) (*domain.ShoppingList, error), idParam, body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShoppingListHandler(&mockShoppingListService{addItemsFn: addItemsFn})
	e.POST("/api/shopping_list/:id/items/bulk/", h.AddItems)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/shopping_list/"+idParam+"/items/bulk/", body))
	return rec
}

// 複数項目を一括追加した時、200 が返ること。
func TestShoppingListHandler_AddItems_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveAddItems(func(_ context.Context, _, _ string, items []service.NewItem) (*domain.ShoppingList, error) {
		return &domain.ShoppingList{ID: testListID, Items: []domain.ShoppingListItem{{ID: testItemID, Name: items[0].Name}}}, nil
	}, testListID, `{"items":[{"name":"牛乳","quantity":200,"unit":"ml"},{"name":"卵"}]}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// items が空配列の時、サービスを呼ばず 400 が返ること。
func TestShoppingListHandler_AddItems_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveAddItems(func(_ context.Context, _, _ string, _ []service.NewItem) (*domain.ShoppingList, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, testListID, `{"items":[]}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// id が UUID でない時、サービスを呼ばず 400 が返ること。
func TestShoppingListHandler_AddItems_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveAddItems(func(_ context.Context, _, _ string, _ []service.NewItem) (*domain.ShoppingList, error) {
		t.Fatal("id 不正時に service を呼んではいけない")
		return nil, nil
	}, "not-a-uuid", `{"items":[{"name":"牛乳"}]}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 他人のリストでサービスが ErrForbidden を返した時、403 が返ること。
func TestShoppingListHandler_AddItems_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveAddItems(func(_ context.Context, _, _ string, _ []service.NewItem) (*domain.ShoppingList, error) {
		return nil, service.ErrForbidden
	}, testListID, `{"items":[{"name":"牛乳"}]}`)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// 存在しないリストでサービスが ErrNotFound を返した時、404 が返ること。
func TestShoppingListHandler_AddItems_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveAddItems(func(_ context.Context, _, _ string, _ []service.NewItem) (*domain.ShoppingList, error) {
		return nil, service.ErrNotFound
	}, testListID, `{"items":[{"name":"牛乳"}]}`)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// serveUpdateItem は setCheckedFn を差し替えたハンドラに PUT /api/shopping_list/:id/items/:item_id/ し結果を返す。
func serveUpdateItem(setCheckedFn func(context.Context, string, string, string, bool) (*domain.ShoppingList, error), idParam, itemParam, body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShoppingListHandler(&mockShoppingListService{setCheckedFn: setCheckedFn})
	e.PUT("/api/shopping_list/:id/items/:item_id/", h.UpdateItem)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPut, "/api/shopping_list/"+idParam+"/items/"+itemParam+"/", body))
	return rec
}

// チェック状態を更新した時、200 が返ること。
func TestShoppingListHandler_UpdateItem_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveUpdateItem(func(_ context.Context, _, _, _ string, checked bool) (*domain.ShoppingList, error) {
		return &domain.ShoppingList{ID: testListID, Items: []domain.ShoppingListItem{{ID: testItemID, Name: "牛乳", Checked: checked}}}, nil
	}, testListID, testItemID, `{"checked":true}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// item_id が UUID でない時、サービスを呼ばず 400 が返ること。
func TestShoppingListHandler_UpdateItem_InvalidItemID(t *testing.T) {
	// Arrange & Act
	rec := serveUpdateItem(func(_ context.Context, _, _, _ string, _ bool) (*domain.ShoppingList, error) {
		t.Fatal("item_id 不正時に service を呼んではいけない")
		return nil, nil
	}, testListID, "not-a-uuid", `{"checked":true}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 存在しない項目でサービスが ErrNotFound を返した時、404 が返ること。
func TestShoppingListHandler_UpdateItem_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveUpdateItem(func(_ context.Context, _, _, _ string, _ bool) (*domain.ShoppingList, error) {
		return nil, service.ErrNotFound
	}, testListID, testItemID, `{"checked":true}`)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// serveDeleteItem は deleteItemFn を差し替えたハンドラに DELETE /api/shopping_list/:id/items/:item_id/ し結果を返す。
func serveDeleteItem(deleteItemFn func(context.Context, string, string, string) (*domain.ShoppingList, error), idParam, itemParam string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShoppingListHandler(&mockShoppingListService{deleteItemFn: deleteItemFn})
	e.DELETE("/api/shopping_list/:id/items/:item_id/", h.DeleteItem)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodDelete, "/api/shopping_list/"+idParam+"/items/"+itemParam+"/", ""))
	return rec
}

// 項目を削除した時、200 が返ること。
func TestShoppingListHandler_DeleteItem_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveDeleteItem(func(_ context.Context, _, _, _ string) (*domain.ShoppingList, error) {
		return &domain.ShoppingList{ID: testListID}, nil
	}, testListID, testItemID)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// serveClearChecked は clearCheckedFn を差し替えたハンドラに DELETE .../items/checked/ し結果を返す。
func serveClearChecked(clearCheckedFn func(context.Context, string, string) (*domain.ShoppingList, error), idParam string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShoppingListHandler(&mockShoppingListService{clearCheckedFn: clearCheckedFn})
	// 静的パス checked を :item_id より先に登録して、本番のルーティングと同じ優先順位にする。
	e.DELETE("/api/shopping_list/:id/items/checked/", h.ClearChecked)
	e.DELETE("/api/shopping_list/:id/items/:item_id/", h.DeleteItem)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodDelete, "/api/shopping_list/"+idParam+"/items/checked/", ""))
	return rec
}

// チェック済みを一括削除した時、200 が返ること(checked が :item_id より優先されること)。
func TestShoppingListHandler_ClearChecked_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveClearChecked(func(_ context.Context, _, _ string) (*domain.ShoppingList, error) {
		return &domain.ShoppingList{ID: testListID}, nil
	}, testListID)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// serveListReorder は reorderFn を差し替えたハンドラに PUT .../items/reorder/ し結果を返す。
func serveListReorder(reorderFn func(context.Context, string, string, []string) (*domain.ShoppingList, error), idParam, body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShoppingListHandler(&mockShoppingListService{reorderFn: reorderFn})
	// 静的パス reorder を :item_id より先に登録して、本番のルーティングと同じ優先順位にする。
	e.PUT("/api/shopping_list/:id/items/reorder/", h.Reorder)
	e.PUT("/api/shopping_list/:id/items/:item_id/", h.UpdateItem)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPut, "/api/shopping_list/"+idParam+"/items/reorder/", body))
	return rec
}

// 並び替えした時、200 が返ること(reorder が :item_id より優先されること)。
func TestShoppingListHandler_Reorder_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveListReorder(func(_ context.Context, _, _ string, _ []string) (*domain.ShoppingList, error) {
		return &domain.ShoppingList{ID: testListID}, nil
	}, testListID, `{"item_ids":["00000000-0000-0000-0000-000000000022"]}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// item_ids に UUID でない値が混じる時、サービスを呼ばず 400 が返ること。
func TestShoppingListHandler_Reorder_InvalidItemID(t *testing.T) {
	// Arrange & Act
	rec := serveListReorder(func(_ context.Context, _, _ string, _ []string) (*domain.ShoppingList, error) {
		t.Fatal("バリデーション失敗時に service を呼んではいけない")
		return nil, nil
	}, testListID, `{"item_ids":["not-a-uuid"]}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 属さない項目でサービスが ErrNotFound を返した時、404 が返ること。
func TestShoppingListHandler_Reorder_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveListReorder(func(_ context.Context, _, _ string, _ []string) (*domain.ShoppingList, error) {
		return nil, service.ErrNotFound
	}, testListID, `{"item_ids":["00000000-0000-0000-0000-000000000022"]}`)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
