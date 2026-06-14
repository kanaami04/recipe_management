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

const testUserID = uint(7)

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
func serveList(t *testing.T, listFn func(context.Context, uint) ([]domain.Recipe, error)) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{listFn: listFn})
	e.GET("/api/recipes/", h.List, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodGet, "/api/recipes/", ""))
	return rec
}

// serveCreate は createFn を差し替えた RecipeHandler に、認証付きで POST /api/recipes/ し結果を返す。
func serveCreate(t *testing.T, createFn func(context.Context, uint, request.RecipeRequest) (*domain.Recipe, error), body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{createFn: createFn})
	e.POST("/api/recipes/", h.Create, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPost, "/api/recipes/", body))
	return rec
}

// serveUpdate は updateFn を差し替えた RecipeHandler に、認証付きで PUT /api/recipes/:id/ し結果を返す。
// idParam は URL に埋め込む :id 部分（不正値の検証にも使う）。
func serveUpdate(t *testing.T, updateFn func(context.Context, uint, uint, request.RecipeRequest) (*domain.Recipe, error), idParam, body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{updateFn: updateFn})
	e.PUT("/api/recipes/:id/", h.Update, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPut, "/api/recipes/"+idParam+"/", body))
	return rec
}

// serveDelete は deleteFn を差し替えた RecipeHandler に、認証付きで DELETE /api/recipes/:id/ し結果を返す。
func serveDelete(t *testing.T, deleteFn func(context.Context, uint, uint) error, idParam string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewRecipeHandler(&mockRecipeService{deleteFn: deleteFn})
	e.DELETE("/api/recipes/:id/", h.Delete, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodDelete, "/api/recipes/"+idParam+"/", ""))
	return rec
}

// recipeList は肉じゃが1件を返す listFn。
func recipeList(_ context.Context, _ uint) ([]domain.Recipe, error) {
	return []domain.Recipe{
		{ID: 1, Title: "肉じゃが", Owner: domain.ApplicationUser{ID: testUserID, Username: "alice"}},
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
	var gotUserID uint
	serveList(t, func(_ context.Context, userID uint) ([]domain.Recipe, error) {
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
		listFn: func(_ context.Context, _ uint) ([]domain.Recipe, error) {
			t.Fatal("認証なしで service を呼んではいけない")
			return nil, nil
		},
	})
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
	rec := serveCreate(t, func(_ context.Context, _ uint, req request.RecipeRequest) (*domain.Recipe, error) {
		return &domain.Recipe{ID: 9, Title: req.Title, Owner: domain.ApplicationUser{ID: testUserID, Username: "alice"}}, nil
	}, `{"title":"カレー"}`)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
}

// 作成した時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_Create_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID uint
	serveCreate(t, func(_ context.Context, userID uint, _ request.RecipeRequest) (*domain.Recipe, error) {
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
	serveCreate(t, func(_ context.Context, _ uint, req request.RecipeRequest) (*domain.Recipe, error) {
		gotReq = req
		return &domain.Recipe{}, nil
	}, `{"title":"カレー","create_for":2}`)

	// Assert
	assert.Equal(t, request.RecipeRequest{Title: "カレー", CreateFor: 2}, gotReq)
}

// 必須項目が欠けている時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Create_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveCreate(t, func(_ context.Context, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, `{}`) // title 欠落

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// okUpdate は受け取ったタイトルで更新後レシピを返す updateFn。
func okUpdate(_ context.Context, _, id uint, req request.RecipeRequest) (*domain.Recipe, error) {
	return &domain.Recipe{ID: id, Title: req.Title}, nil
}

// 認証済みで更新した時、200 が返ること。
func TestRecipeHandler_Update_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, okUpdate, "5", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 更新した時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_Update_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID uint
	serveUpdate(t, func(_ context.Context, userID, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		gotUserID = userID
		return &domain.Recipe{}, nil
	}, "5", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, testUserID, gotUserID)
}

// 更新した時、URL の :id がサービスに渡されること。
func TestRecipeHandler_Update_ForwardsRecipeID(t *testing.T) {
	// Arrange & Act
	var gotID uint
	serveUpdate(t, func(_ context.Context, _, recipeID uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		gotID = recipeID
		return &domain.Recipe{}, nil
	}, "5", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, uint(5), gotID)
}

// 対象レシピが存在しない時、404 が返ること。
func TestRecipeHandler_Update_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		return nil, service.ErrNotFound
	}, "5", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// 権限が無いユーザーが更新した時、403 が返ること。
func TestRecipeHandler_Update_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		return nil, service.ErrForbidden
	}, "5", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// 共有先ユーザーが存在しない時、400 が返ること。
func TestRecipeHandler_Update_SharedUserNotFound(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		return nil, service.ErrSharedUserNotFound
	}, "5", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// サービスが想定外のエラーを返した時、500 が返ること。
func TestRecipeHandler_Update_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		return nil, assert.AnError
	}, "5", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// :id が数値でない時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Update_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		t.Fatal("不正な id では service を呼んではいけない")
		return nil, nil
	}, "abc", `{"title":"更新"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 必須項目が欠けている時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Update_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveUpdate(t, func(_ context.Context, _, _ uint, _ request.RecipeRequest) (*domain.Recipe, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, "5", `{}`) // title 欠落

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 認証済みで削除した時、204 が返ること。
func TestRecipeHandler_Delete_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveDelete(t, func(_ context.Context, _, _ uint) error {
		return nil
	}, "5")

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// 削除した時、JWT のユーザーIDがサービスに渡されること。
func TestRecipeHandler_Delete_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotUserID uint
	serveDelete(t, func(_ context.Context, userID, _ uint) error {
		gotUserID = userID
		return nil
	}, "5")

	// Assert
	assert.Equal(t, testUserID, gotUserID)
}

// 削除した時、URL の :id がサービスに渡されること。
func TestRecipeHandler_Delete_ForwardsRecipeID(t *testing.T) {
	// Arrange & Act
	var gotID uint
	serveDelete(t, func(_ context.Context, _, recipeID uint) error {
		gotID = recipeID
		return nil
	}, "5")

	// Assert
	assert.Equal(t, uint(5), gotID)
}

// 対象レシピが存在しない時、404 が返ること。
func TestRecipeHandler_Delete_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveDelete(t, func(_ context.Context, _, _ uint) error {
		return service.ErrNotFound
	}, "5")

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// 権限が無いユーザーが削除した時、403 が返ること。
func TestRecipeHandler_Delete_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveDelete(t, func(_ context.Context, _, _ uint) error {
		return service.ErrForbidden
	}, "5")

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// :id が数値でない時、サービスを呼ばず 400 が返ること。
func TestRecipeHandler_Delete_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveDelete(t, func(_ context.Context, _, _ uint) error {
		t.Fatal("不正な id では service を呼んではいけない")
		return nil
	}, "abc")

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
