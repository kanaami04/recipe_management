package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"recipe-backend/internal/domain"
	appmw "recipe-backend/internal/middleware"
	jwtpkg "recipe-backend/internal/pkg/jwt"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
)

// serveUserInfo は getByIDFn を差し替えた UserHandler に、認証付きで GET /api/user_info/ し結果を返す。
func serveUserInfo(t *testing.T, getByIDFn func(context.Context, uint) (*domain.User, error)) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{getByIDFn: getByIDFn})
	e.GET("/api/user_info/", h.Info, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodGet, "/api/user_info/", ""))
	return rec
}

// serveUserList は listFn を差し替えた UserHandler に GET /api/users/ し結果を返す。
func serveUserList(listFn func(context.Context) ([]domain.User, error)) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{listFn: listFn})
	e.GET("/api/users/", h.List)
	req := httptest.NewRequest(http.MethodGet, "/api/users/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// infoUser は testUserID の有効ユーザーを返す getByIDFn。
func infoUser(_ context.Context, _ uint) (*domain.User, error) {
	return factory.NewUser(factory.WithID(testUserID), factory.WithUsername("alice")), nil
}

// 認証済みでユーザー情報を取得した時、200 が返ること。
func TestUserHandler_Info_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveUserInfo(t, infoUser)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 認証済みでユーザー情報を取得した時、レスポンスにユーザー名が含まれること。
func TestUserHandler_Info_ReturnsUserInBody(t *testing.T) {
	// Arrange & Act
	rec := serveUserInfo(t, infoUser)

	// Assert
	assert.Contains(t, rec.Body.String(), "alice")
}

// ユーザー情報を取得した時、JWT のユーザーIDがサービスに渡されること。
func TestUserHandler_Info_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotID uint
	serveUserInfo(t, func(_ context.Context, id uint) (*domain.User, error) {
		gotID = id
		return factory.NewUser(factory.WithID(testUserID)), nil
	})

	// Assert
	assert.Equal(t, testUserID, gotID)
}

// 該当ユーザーが存在しない時、404 が返ること。
func TestUserHandler_Info_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveUserInfo(t, func(_ context.Context, _ uint) (*domain.User, error) {
		return nil, nil
	})

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// サービスがエラーを返した時、500 が返ること。
func TestUserHandler_Info_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveUserInfo(t, func(_ context.Context, _ uint) (*domain.User, error) {
		return nil, assert.AnError
	})

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ユーザー一覧を取得した時、200 が返ること。
func TestUserHandler_List_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveUserList(func(_ context.Context) ([]domain.User, error) {
		return []domain.User{*factory.NewUser(factory.WithID(1), factory.WithUsername("alice"))}, nil
	})

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ユーザー一覧を取得した時、レスポンスにユーザー名が含まれること。
func TestUserHandler_List_ReturnsUsersInBody(t *testing.T) {
	// Arrange & Act
	rec := serveUserList(func(_ context.Context) ([]domain.User, error) {
		return []domain.User{*factory.NewUser(factory.WithID(1), factory.WithUsername("alice"))}, nil
	})

	// Assert
	assert.Contains(t, rec.Body.String(), "alice")
}

// サービスがエラーを返した時、500 が返ること。
func TestUserHandler_List_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveUserList(func(_ context.Context) ([]domain.User, error) {
		return nil, assert.AnError
	})

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
