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

const testGroupMemberID = "00000000-0000-0000-0000-000000000031"

func newShareGroupHandler(svc *mockShareGroupService) *ShareGroupHandler {
	return NewShareGroupHandler(svc, mockAvatarStorage{})
}

func sampleGroup() *domain.ShareGroup {
	return &domain.ShareGroup{
		ID:      "00000000-0000-0000-0000-000000000030",
		Name:    "我が家",
		OwnerID: "u1",
		Owner:   domain.User{ID: "u1", Username: "taro"},
		Members: []domain.User{{ID: "u1", Username: "taro"}},
	}
}

// serveGroupGet は getMineFn を差し替えたハンドラに GET /api/share_group/ し結果を返す。
func serveGroupGet(getMineFn func(context.Context, string) (*domain.ShareGroup, error)) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShareGroupHandler(&mockShareGroupService{getMineFn: getMineFn})
	e.GET("/api/share_group/", h.Get)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/share_group/", nil))
	return rec
}

// グループを取得した時、200 が返ること。
func TestShareGroupHandler_Get_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveGroupGet(func(_ context.Context, _ string) (*domain.ShareGroup, error) {
		return sampleGroup(), nil
	})

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 未所属のとき Get で 404 が返ること。
func TestShareGroupHandler_Get_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveGroupGet(func(_ context.Context, _ string) (*domain.ShareGroup, error) {
		return nil, nil
	})

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// serveGroupCreate は createFn を差し替えたハンドラに POST /api/share_group/ し結果を返す。
func serveGroupCreate(createFn func(context.Context, string, string) (*domain.ShareGroup, error), body string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShareGroupHandler(&mockShareGroupService{createFn: createFn})
	e.POST("/api/share_group/", h.Create)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/share_group/", body))
	return rec
}

// グループを作成した時、201 が返ること。
func TestShareGroupHandler_Create_Returns201(t *testing.T) {
	// Arrange & Act
	rec := serveGroupCreate(func(_ context.Context, _, _ string) (*domain.ShareGroup, error) {
		return sampleGroup(), nil
	}, `{"name":"我が家"}`)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
}

// 既に所属していてサービスが ErrAlreadyInGroup を返した時、409 が返ること。
func TestShareGroupHandler_Create_Conflict(t *testing.T) {
	// Arrange & Act
	rec := serveGroupCreate(func(_ context.Context, _, _ string) (*domain.ShareGroup, error) {
		return nil, service.ErrAlreadyInGroup
	}, `{"name":"我が家"}`)

	// Assert
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// serveGroupJoin は joinFn を差し替えたハンドラに POST /api/share_group/join/ し結果を返す。
func serveGroupJoin(t *testing.T, joinFn func(context.Context, string, string) (*domain.ShareGroup, error), body string) *httptest.ResponseRecorder {
	t.Helper()
	e := newTestEcho()
	h := newShareGroupHandler(&mockShareGroupService{joinFn: joinFn})
	e.POST("/api/share_group/join/", h.Join)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/share_group/join/", body))
	return rec
}

// 招待コードで参加した時、200 が返ること。
func TestShareGroupHandler_Join_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveGroupJoin(t, func(_ context.Context, _, _ string) (*domain.ShareGroup, error) {
		return sampleGroup(), nil
	}, `{"invite_code":"CODE1234"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// invite_code が空の時、サービスを呼ばず 400 が返ること。
func TestShareGroupHandler_Join_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveGroupJoin(t, func(_ context.Context, _, _ string) (*domain.ShareGroup, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, `{"invite_code":""}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 無効な招待コードでサービスが ErrInviteCodeInvalid を返した時、400 が返ること。
func TestShareGroupHandler_Join_InvalidCode(t *testing.T) {
	// Arrange & Act
	rec := serveGroupJoin(t, func(_ context.Context, _, _ string) (*domain.ShareGroup, error) {
		return nil, service.ErrInviteCodeInvalid
	}, `{"invite_code":"NOPE"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// serveGroupLeave は leaveFn を差し替えたハンドラに POST /api/share_group/leave/ し結果を返す。
func serveGroupLeave(leaveFn func(context.Context, string) error) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShareGroupHandler(&mockShareGroupService{leaveFn: leaveFn})
	e.POST("/api/share_group/leave/", h.Leave)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/share_group/leave/", ""))
	return rec
}

// グループを抜けた時、204 が返ること。
func TestShareGroupHandler_Leave_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveGroupLeave(func(_ context.Context, _ string) error {
		return nil
	})

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// 未所属でサービスが ErrNotInGroup を返した時、404 が返ること。
func TestShareGroupHandler_Leave_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveGroupLeave(func(_ context.Context, _ string) error {
		return service.ErrNotInGroup
	})

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// serveGroupRegenerate は regenerateFn を差し替えたハンドラに POST .../invite_code/ し結果を返す。
func serveGroupRegenerate(regenerateFn func(context.Context, string) (*domain.ShareGroup, error)) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShareGroupHandler(&mockShareGroupService{regenerateFn: regenerateFn})
	e.POST("/api/share_group/invite_code/", h.RegenerateInviteCode)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/share_group/invite_code/", ""))
	return rec
}

// 招待コードを再発行した時、200 が返ること。
func TestShareGroupHandler_Regenerate_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveGroupRegenerate(func(_ context.Context, _ string) (*domain.ShareGroup, error) {
		return sampleGroup(), nil
	})

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 所有者でなくサービスが ErrNotGroupOwner を返した時、403 が返ること。
func TestShareGroupHandler_Regenerate_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveGroupRegenerate(func(_ context.Context, _ string) (*domain.ShareGroup, error) {
		return nil, service.ErrNotGroupOwner
	})

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// serveRemoveMember は removeFn を差し替えたハンドラに DELETE .../members/:user_id/ し結果を返す。
func serveRemoveMember(removeFn func(context.Context, string, string) error, userIDParam string) *httptest.ResponseRecorder {
	e := newTestEcho()
	h := newShareGroupHandler(&mockShareGroupService{removeFn: removeFn})
	e.DELETE("/api/share_group/members/:user_id/", h.RemoveMember)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, jsonRequest(http.MethodDelete, "/api/share_group/members/"+userIDParam+"/", ""))
	return rec
}

// メンバーを外した時、204 が返ること。
func TestShareGroupHandler_RemoveMember_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveRemoveMember(func(_ context.Context, _, _ string) error {
		return nil
	}, testGroupMemberID)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// user_id が UUID でない時、サービスを呼ばず 400 が返ること。
func TestShareGroupHandler_RemoveMember_InvalidID(t *testing.T) {
	// Arrange & Act
	rec := serveRemoveMember(func(_ context.Context, _, _ string) error {
		t.Fatal("id 不正時に service を呼んではいけない")
		return nil
	}, "not-a-uuid")

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 所有者でなくサービスが ErrNotGroupOwner を返した時、403 が返ること。
func TestShareGroupHandler_RemoveMember_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveRemoveMember(func(_ context.Context, _, _ string) error {
		return service.ErrNotGroupOwner
	}, testGroupMemberID)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
