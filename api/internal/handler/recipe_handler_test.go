package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
	appmw "recipe-backend/internal/middleware"
	jwtpkg "recipe-backend/internal/pkg/jwt"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID   = "00000000-0000-0000-0000-000000000007"
	testRecipeID = "00000000-0000-0000-0000-000000000005"
)

// authedRequest は testUserID 用のアクセストークンを付けたリクエストを作る。
func authedRequest(t *testing.T, jm *jwtpkg.Manager, method, path, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	token, err := jm.GenerateAccess(testUserID)
	require.NoError(t, err)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	return req
}

// serveList は listFn を差し替えた RecipeHandler に、認証付きで GET /api/recipes/ し結果を返す。
func serveList(t *testing.T, listFn func(context.Context, string) ([]domain.Recipe, error)) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{listFn: listFn}, mockAvatarStorage{})
	e.GET("/api/recipes/", h.List, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodGet, "/api/recipes/", ""))
	return rec
}

// serveCreate は createFn を差し替えた RecipeHandler に、認証付きで POST /api/recipes/ し結果を返す。
func serveCreate(t *testing.T, createFn func(context.Context, string, request.RecipeRequest) (*domain.Recipe, error), body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{createFn: createFn}, mockAvatarStorage{})
	e.POST("/api/recipes/", h.Create, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPost, "/api/recipes/", body))
	return rec
}

// serveUpdate は updateFn を差し替えた RecipeHandler に、認証付きで PUT /api/recipes/:id/ し結果を返す。
// idParam は URL に埋め込む :id 部分（不正値の検証にも使う）。
func serveUpdate(t *testing.T, updateFn func(context.Context, string, string, request.RecipeRequest) (*domain.Recipe, error), idParam, body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{updateFn: updateFn}, mockAvatarStorage{})
	e.PUT("/api/recipes/:id/", h.Update, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPut, "/api/recipes/"+idParam+"/", body))
	return rec
}

// serveDelete は deleteFn を差し替えた RecipeHandler に、認証付きで DELETE /api/recipes/:id/ し結果を返す。
func serveDelete(t *testing.T, deleteFn func(context.Context, string, string) error, idParam string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{deleteFn: deleteFn}, mockAvatarStorage{})
	e.DELETE("/api/recipes/:id/", h.Delete, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodDelete, "/api/recipes/"+idParam+"/", ""))
	return rec
}

// recipeList は肉じゃが1件を返す listFn。
func recipeList(_ context.Context, _ string) ([]domain.Recipe, error) {
	return []domain.Recipe{
		{ID: "r1", Title: "肉じゃが", Owner: domain.User{ID: testUserID, Username: "alice"}},
	}, nil
}

// 認証済みで一覧取得した時、200 が返ること。
func TestRecipeHandler_List_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveList(t, recipeList)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 認証済みで一覧取得した時、レスポンスにレシピが含まれること。
func TestRecipeHandler_List_ReturnsRecipesInBody(t *testing.T) {
	// Arrange & Act
	rec := serveList(t, recipeList)

	// Assert
	assert.Contains(t, rec.Body.String(), "肉じゃが")
}

// 一覧取得した時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_List_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID string
	serveList(t, func(_ context.Context, userID string) ([]domain.Recipe, error) {
		gotUserID = userID
		return nil, nil
	})

	// Assert
	assert.Equal(t, testUserID, gotUserID)
}

// トークンなしで一覧取得した時、サービスを呼ばず 401 が返ること。
func TestRecipeHandler_List_Unauthorized(t *testing.T) {
	// Arrange
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{
		listFn: func(_ context.Context, _ string) ([]domain.Recipe, error) {
			t.Fatal("認証なしで service を呼んではいけない")
			return nil, nil
		},
	}, mockAvatarStorage{})
	e.GET("/api/recipes/", h.List, appmw.JWT(jm))

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/recipes/", nil) // トークンなし
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// 認証済みで作成した時、201 が返ること。
func TestRecipeHandler_Create_Returns201(t *testing.T) {
	// Arrange & Act
	rec := serveCreate(t, func(_ context.Context, _ string, req request.RecipeRequest) (*domain.Recipe, error) {
		return &domain.Recipe{ID: "r9", Title: req.Title, Owner: domain.User{ID: testUserID, Username: "alice"}}, nil
	}, `{"title":"カレー"}`)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
}

// 作成した時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_Create_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID string
	serveCreate(t, func(_ context.Context, userID string, _ request.RecipeRequest) (*domain.Recipe, error) {
		gotUserID = userID
		return &domain.Recipe{}, nil
	}, `{"title":"カレー"}`)

	// Assert
	assert.Equal(t, testUserID, gotUserID)
}

// 作成した時、パースされたリクエストが構造体ごとサービスに渡されること。
func TestRecipeHandler_Create_ForwardsParsedRequest(t *testing.T) {
	// Arrange & Act
	var gotReq request.RecipeRequest
	serveCreate(t, func(_ context.Context, _ string, req request.RecipeRequest) (*domain.Recipe, error) {
		gotReq = req
		return &domain.Recipe{}, nil
	}, `{"title":"カレー","create_for":2}`)

	// Assert
	assert.Equal(t, request.RecipeRequest{Title: "カレー", CreateFor: 2}, gotReq)
}

// 必須項目が欠けている時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Create_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveCreate(t, func(_ context.Context, _ string, _ request.RecipeRequest) (*domain.Recipe, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, `{}`) // title 欠落

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// okUpdate は受け取ったタイトルで更新後レシピを返す updateFn。
func okUpdate(_ context.Context, _, id string, req request.RecipeRequest) (*domain.Recipe, error) {
	return &domain.Recipe{ID: id, Title: req.Title}, nil
}

// 認証済みで更新した時、200 が返ること。
func TestRecipeHandler_Update_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, okUpdate, testRecipeID, `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 更新した時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_Update_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID string
	serveUpdate(t, func(_ context.Context, userID, _ string, _ request.RecipeRequest) (*domain.Recipe, error) {
		gotUserID = userID
		return &domain.Recipe{}, nil
	}, testRecipeID, `{"title":"更新"}`)

	// Assert
	assert.Equal(t, testUserID, gotUserID)
}

// 更新した時、URL の :id がサービスに渡されること。
func TestRecipeHandler_Update_ForwardsRecipeID(t *testing.T) {
	// Arrange & Act
	var gotID string
	serveUpdate(t, func(_ context.Context, _, recipeID string, _ request.RecipeRequest) (*domain.Recipe, error) {
		gotID = recipeID
		return &domain.Recipe{}, nil
	}, testRecipeID, `{"title":"更新"}`)

	// Assert
	assert.Equal(t, testRecipeID, gotID)
}

// 対象レシピが存在しない時、404 が返ること。
func TestRecipeHandler_Update_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ string, _ request.RecipeRequest) (*domain.Recipe, error) {
		return nil, service.ErrNotFound
	}, testRecipeID, `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// 権限が無いユーザーが更新した時、403 が返ること。
func TestRecipeHandler_Update_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ string, _ request.RecipeRequest) (*domain.Recipe, error) {
		return nil, service.ErrForbidden
	}, testRecipeID, `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// 共有先ユーザーが存在しない時、400 が返ること。
func TestRecipeHandler_Update_SharedUserNotFound(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ string, _ request.RecipeRequest) (*domain.Recipe, error) {
		return nil, service.ErrSharedUserNotFound
	}, testRecipeID, `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// サービスが想定外のエラーを返した時、500 が返ること。
func TestRecipeHandler_Update_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ string, _ request.RecipeRequest) (*domain.Recipe, error) {
		return nil, assert.AnError
	}, testRecipeID, `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// :id が数値でない時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Update_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ string, _ request.RecipeRequest) (*domain.Recipe, error) {
		t.Fatal("不正な id では service を呼んではいけない")
		return nil, nil
	}, "abc", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 必須項目が欠けている時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Update_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ string, _ request.RecipeRequest) (*domain.Recipe, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, testRecipeID, `{}`) // title 欠落

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 認証済みで削除した時、204 が返ること。
func TestRecipeHandler_Delete_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveDelete(t, func(_ context.Context, _, _ string) error {
		return nil
	}, testRecipeID)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// 削除した時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_Delete_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID string
	serveDelete(t, func(_ context.Context, userID, _ string) error {
		gotUserID = userID
		return nil
	}, testRecipeID)

	// Assert
	assert.Equal(t, testUserID, gotUserID)
}

// 削除した時、URL の :id がサービスに渡されること。
func TestRecipeHandler_Delete_ForwardsRecipeID(t *testing.T) {
	// Arrange & Act
	var gotID string
	serveDelete(t, func(_ context.Context, _, recipeID string) error {
		gotID = recipeID
		return nil
	}, testRecipeID)

	// Assert
	assert.Equal(t, testRecipeID, gotID)
}

// 対象レシピが存在しない時、404 が返ること。
func TestRecipeHandler_Delete_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveDelete(t, func(_ context.Context, _, _ string) error {
		return service.ErrNotFound
	}, testRecipeID)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// 権限が無いユーザーが削除した時、403 が返ること。
func TestRecipeHandler_Delete_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveDelete(t, func(_ context.Context, _, _ string) error {
		return service.ErrForbidden
	}, testRecipeID)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// :id が数値でない時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Delete_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveDelete(t, func(_ context.Context, _, _ string) error {
		t.Fatal("不正な id では service を呼んではいけない")
		return nil
	}, "abc")

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// serveReorder は reorderFn を差し替えた RecipeHandler に、認証付きで PUT /api/recipes/reorder/ し結果を返す。
func serveReorder(t *testing.T, reorderFn func(context.Context, string, []string) error, body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{reorderFn: reorderFn}, mockAvatarStorage{})
	e.PUT("/api/recipes/reorder/", h.Reorder, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPut, "/api/recipes/reorder/", body))
	return rec
}

// 認証済みで並び替えした時、204 が返ること。
func TestRecipeHandler_Reorder_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveReorder(t, func(_ context.Context, _ string, _ []string) error {
		return nil
	}, `{"recipe_ids":["`+testRecipeID+`"]}`)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// 並び替えした時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_Reorder_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID string
	serveReorder(t, func(_ context.Context, userID string, _ []string) error {
		gotUserID = userID
		return nil
	}, `{"recipe_ids":["`+testRecipeID+`"]}`)

	// Assert
	assert.Equal(t, testUserID, gotUserID)
}

// 並び替えした時、リクエストの並び順がサービスに渡されること。
func TestRecipeHandler_Reorder_ForwardsOrder(t *testing.T) {
	// Arrange & Act
	var gotIDs []string
	serveReorder(t, func(_ context.Context, _ string, recipeIDs []string) error {
		gotIDs = recipeIDs
		return nil
	}, `{"recipe_ids":["`+testRecipeID+`","`+testUserID+`"]}`)

	// Assert
	assert.Equal(t, []string{testRecipeID, testUserID}, gotIDs)
}

// recipe_ids が UUID でない時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Reorder_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveReorder(t, func(_ context.Context, _ string, _ []string) error {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil
	}, `{"recipe_ids":["not-a-uuid"]}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 閲覧できないレシピを含む並び替えでサービスが ErrForbidden を返した時、403 が返ること。
func TestRecipeHandler_Reorder_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveReorder(t, func(_ context.Context, _ string, _ []string) error {
		return service.ErrForbidden
	}, `{"recipe_ids":["`+testRecipeID+`"]}`)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// サービスが想定外のエラーを返した時、500 が返ること。
func TestRecipeHandler_Reorder_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveReorder(t, func(_ context.Context, _ string, _ []string) error {
		return assert.AnError
	}, `{"recipe_ids":["`+testRecipeID+`"]}`)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// serveArchive は archiveFn を差し替えた RecipeHandler に、認証付きで PUT /api/recipes/:id/archive/ し結果を返す。
func serveArchive(t *testing.T, archiveFn func(context.Context, string, string, bool) error, idParam, body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{archiveFn: archiveFn}, mockAvatarStorage{})
	e.PUT("/api/recipes/:id/archive/", h.Archive, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPut, "/api/recipes/"+idParam+"/archive/", body))
	return rec
}

// 認証済みでアーカイブ更新した時、204 が返ること。
func TestRecipeHandler_Archive_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveArchive(t, func(_ context.Context, _, _ string, _ bool) error {
		return nil
	}, testRecipeID, `{"archive_flg":true}`)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// アーカイブ更新した時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_Archive_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID string
	serveArchive(t, func(_ context.Context, userID, _ string, _ bool) error {
		gotUserID = userID
		return nil
	}, testRecipeID, `{"archive_flg":true}`)

	// Assert
	assert.Equal(t, testUserID, gotUserID)
}

// アーカイブ更新した時、対象レシピIDと状態がサービスに渡されること。
func TestRecipeHandler_Archive_ForwardsTarget(t *testing.T) {
	// Arrange & Act
	var gotRecipeID string
	var gotArchived bool
	serveArchive(t, func(_ context.Context, _, recipeID string, archived bool) error {
		gotRecipeID, gotArchived = recipeID, archived
		return nil
	}, testRecipeID, `{"archive_flg":true}`)

	// Assert
	assert.Equal(t, testRecipeID, gotRecipeID)
	assert.True(t, gotArchived)
}

// id が UUID でない時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Archive_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveArchive(t, func(_ context.Context, _, _ string, _ bool) error {
		t.Fatal("id 不正時に service を呼んではいけない")
		return nil
	}, "not-a-uuid", `{"archive_flg":true}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 閲覧できないレシピのアーカイブでサービスが ErrForbidden を返した時、403 が返ること。
func TestRecipeHandler_Archive_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveArchive(t, func(_ context.Context, _, _ string, _ bool) error {
		return service.ErrForbidden
	}, testRecipeID, `{"archive_flg":true}`)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// 存在しないレシピのアーカイブでサービスが ErrNotFound を返した時、404 が返ること。
func TestRecipeHandler_Archive_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveArchive(t, func(_ context.Context, _, _ string, _ bool) error {
		return service.ErrNotFound
	}, testRecipeID, `{"archive_flg":false}`)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// サービスが想定外のエラーを返した時、500 が返ること。
func TestRecipeHandler_Archive_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveArchive(t, func(_ context.Context, _, _ string, _ bool) error {
		return assert.AnError
	}, testRecipeID, `{"archive_flg":true}`)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
