package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"recipe-backend/internal/domain"
	appmw "recipe-backend/internal/middleware"
	jwtpkg "recipe-backend/internal/pkg/jwt"
	"recipe-backend/internal/service"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
)

// serveUserInfo は getByIDFn を差し替えた UserHandler に、認証付きで GET /api/user_info/ し結果を返す。
func serveUserInfo(t *testing.T, getByIDFn func(context.Context, string) (*domain.User, error)) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{getByIDFn: getByIDFn}, mockAvatarStorage{})
	e.GET("/api/user_info/", h.Info, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodGet, "/api/user_info/", ""))
	return rec
}

// infoUser は testUserID の有効ユーザーを返す getByIDFn。
func infoUser(_ context.Context, _ string) (*domain.User, error) {
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
	var gotID string
	serveUserInfo(t, func(_ context.Context, id string) (*domain.User, error) {
		gotID = id
		return factory.NewUser(factory.WithID(testUserID)), nil
	})

	// Assert
	assert.Equal(t, testUserID, gotID)
}

// 該当ユーザーが存在しない時、404 が返ること。
func TestUserHandler_Info_NotFound(t *testing.T) {
	// Arrange & Act
	rec := serveUserInfo(t, func(_ context.Context, _ string) (*domain.User, error) {
		return nil, nil
	})

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// サービスがエラーを返した時、500 が返ること。
func TestUserHandler_Info_InternalError(t *testing.T) {
	// Arrange & Act
	rec := serveUserInfo(t, func(_ context.Context, _ string) (*domain.User, error) {
		return nil, assert.AnError
	})

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// serveUserUpdate は updateFn を差し替えた UserHandler に、認証付きで PUT /api/user_info/ し結果を返す。
func serveUserUpdate(t *testing.T, updateFn func(context.Context, string, string) (*domain.User, error), body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{updateFn: updateFn}, mockAvatarStorage{})
	e.PUT("/api/user_info/", h.Update, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPut, "/api/user_info/", body))
	return rec
}

// プロフィール更新した時、200 が返ること。
func TestUserHandler_Update_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveUserUpdate(t, func(_ context.Context, _, username string) (*domain.User, error) {
		return factory.NewUser(factory.WithID(testUserID), factory.WithUsername(username)), nil
	}, `{"username":"alice2"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 必須項目が欠けている時、サービスを呼ばず 400 が返ること。
func TestUserHandler_Update_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveUserUpdate(t, func(_ context.Context, _, _ string) (*domain.User, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, `{}`) // username 欠落

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 別ユーザーと重複しサービスが ErrUserAlreadyExists を返した時、409 が返ること。
func TestUserHandler_Update_Conflict(t *testing.T) {
	// Arrange & Act
	rec := serveUserUpdate(t, func(_ context.Context, _, _ string) (*domain.User, error) {
		return nil, service.ErrUserAlreadyExists
	}, `{"username":"bob"}`)

	// Assert
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// serveChangeEmail は changeEmailFn を差し替えた UserHandler に、認証付きで PUT /api/user_info/email/ し結果を返す。
func serveChangeEmail(t *testing.T, changeEmailFn func(context.Context, string, string, string) (*domain.User, error), body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{changeEmailFn: changeEmailFn}, mockAvatarStorage{})
	e.PUT("/api/user_info/email/", h.ChangeEmail, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPut, "/api/user_info/email/", body))
	return rec
}

// メール変更した時、200 が返ること。
func TestUserHandler_ChangeEmail_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveChangeEmail(t, func(_ context.Context, _, email, _ string) (*domain.User, error) {
		return factory.NewUser(factory.WithID(testUserID), factory.WithEmail(email)), nil
	}, `{"email":"alice2@example.com","password":"pass1234"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// パスワードが欠けている時、サービスを呼ばず 400 が返ること。
func TestUserHandler_ChangeEmail_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveChangeEmail(t, func(_ context.Context, _, _, _ string) (*domain.User, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil, nil
	}, `{"email":"alice2@example.com"}`) // password 欠落

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// パスワードが違いサービスが ErrIncorrectPassword を返した時、400 が返ること。
func TestUserHandler_ChangeEmail_WrongPassword(t *testing.T) {
	// Arrange & Act
	rec := serveChangeEmail(t, func(_ context.Context, _, _, _ string) (*domain.User, error) {
		return nil, service.ErrIncorrectPassword
	}, `{"email":"alice2@example.com","password":"wrong"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 別ユーザーと重複しサービスが ErrUserAlreadyExists を返した時、409 が返ること。
func TestUserHandler_ChangeEmail_Conflict(t *testing.T) {
	// Arrange & Act
	rec := serveChangeEmail(t, func(_ context.Context, _, _, _ string) (*domain.User, error) {
		return nil, service.ErrUserAlreadyExists
	}, `{"email":"bob@example.com","password":"pass1234"}`)

	// Assert
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// serveChangePassword は changePwFn を差し替えた UserHandler に、認証付きで PUT /api/user_info/password/ し結果を返す。
func serveChangePassword(t *testing.T, changePwFn func(context.Context, string, string, string) error, body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{changePwFn: changePwFn}, mockAvatarStorage{})
	e.PUT("/api/user_info/password/", h.ChangePassword, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPut, "/api/user_info/password/", body))
	return rec
}

// パスワード変更した時、204 が返ること。
func TestUserHandler_ChangePassword_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveChangePassword(t, func(_ context.Context, _, _, _ string) error {
		return nil
	}, `{"current_password":"oldpass12","new_password":"newpass34"}`)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// 新パスワードが短すぎる時、サービスを呼ばず 400 が返ること。
func TestUserHandler_ChangePassword_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveChangePassword(t, func(_ context.Context, _, _, _ string) error {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return nil
	}, `{"current_password":"oldpass12","new_password":"short"}`) // 8 文字未満

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 現在のパスワードが違いサービスが ErrIncorrectPassword を返した時、400 が返ること。
func TestUserHandler_ChangePassword_WrongCurrent(t *testing.T) {
	// Arrange & Act
	rec := serveChangePassword(t, func(_ context.Context, _, _, _ string) error {
		return service.ErrIncorrectPassword
	}, `{"current_password":"wrong","new_password":"newpass34"}`)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// serveDeleteAccount は deleteFn を差し替えた UserHandler に、認証付きで DELETE /api/user_info/ し結果を返す。
func serveDeleteAccount(t *testing.T, deleteFn func(context.Context, string) error) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{deleteFn: deleteFn}, mockAvatarStorage{})
	e.DELETE("/api/user_info/", h.Delete, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodDelete, "/api/user_info/", ""))
	return rec
}

// アカウント削除した時、204 が返ること。
func TestUserHandler_Delete_Returns204(t *testing.T) {
	// Arrange & Act
	rec := serveDeleteAccount(t, func(_ context.Context, _ string) error {
		return nil
	})

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// アカウント削除した時、JWT のユーザーIDがサービスに渡されること。
func TestUserHandler_Delete_ForwardsUserID(t *testing.T) {
	// Arrange & Act
	var gotID string
	serveDeleteAccount(t, func(_ context.Context, userID string) error {
		gotID = userID
		return nil
	})

	// Assert
	assert.Equal(t, testUserID, gotID)
}

// serveCreateAvatar は createAvatarFn を差し替えた UserHandler に POST /api/user_info/avatar/ し結果を返す。
func serveCreateAvatar(t *testing.T, createAvatarFn func(context.Context, string, string) (string, string, error), body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{createAvatarFn: createAvatarFn}, mockAvatarStorage{})
	e.POST("/api/user_info/avatar/", h.CreateAvatarUploadURL, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPost, "/api/user_info/avatar/", body))
	return rec
}

// アップロード URL 発行した時、200 が返ること。
func TestUserHandler_CreateAvatar_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveCreateAvatar(t, func(_ context.Context, _, _ string) (string, string, error) {
		return "https://s3/upload", "avatars/u/abc", nil
	}, `{"content_type":"image/png"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// アップロード URL 発行した時、レスポンスにアップロード先 URL が含まれること。
func TestUserHandler_CreateAvatar_ReturnsUploadURL(t *testing.T) {
	// Arrange & Act
	rec := serveCreateAvatar(t, func(_ context.Context, _, _ string) (string, string, error) {
		return "https://s3/upload", "avatars/u/abc", nil
	}, `{"content_type":"image/png"}`)

	// Assert
	assert.Contains(t, rec.Body.String(), "https://s3/upload")
}

// content_type が許可外の時、サービスを呼ばず 400 が返ること。
func TestUserHandler_CreateAvatar_ValidationError(t *testing.T) {
	// Arrange & Act
	rec := serveCreateAvatar(t, func(_ context.Context, _, _ string) (string, string, error) {
		t.Fatal("validation fail 時に service を呼んではいけない")
		return "", "", nil
	}, `{"content_type":"image/gif"}`) // 許可外

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// serveConfirmAvatar は confirmAvatarFn を差し替えた UserHandler に PUT /api/user_info/avatar/ し結果を返す。
func serveConfirmAvatar(t *testing.T, confirmAvatarFn func(context.Context, string, string) (*domain.User, error), body string) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{confirmAvatarFn: confirmAvatarFn}, mockAvatarStorage{})
	e.PUT("/api/user_info/avatar/", h.ConfirmAvatar, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodPut, "/api/user_info/avatar/", body))
	return rec
}

// アバターを確定した時、200 が返ること。
func TestUserHandler_ConfirmAvatar_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveConfirmAvatar(t, func(_ context.Context, _, _ string) (*domain.User, error) {
		return factory.NewUser(factory.WithID(testUserID)), nil
	}, `{"key":"avatars/u/abc"}`)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// 他人の key でサービスが ErrForbidden を返した時、403 が返ること。
func TestUserHandler_ConfirmAvatar_Forbidden(t *testing.T) {
	// Arrange & Act
	rec := serveConfirmAvatar(t, func(_ context.Context, _, _ string) (*domain.User, error) {
		return nil, service.ErrForbidden
	}, `{"key":"avatars/other/abc"}`)

	// Assert
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// serveDeleteAvatar は deleteAvatarFn を差し替えた UserHandler に DELETE /api/user_info/avatar/ し結果を返す。
func serveDeleteAvatar(t *testing.T, deleteAvatarFn func(context.Context, string) (*domain.User, error)) *httptest.ResponseRecorder {
	t.Helper()
	jm := jwtpkg.NewManager("secret")
	e := newTestEcho()
	h := NewUserHandler(&mockUserService{deleteAvatarFn: deleteAvatarFn}, mockAvatarStorage{})
	e.DELETE("/api/user_info/avatar/", h.DeleteAvatar, appmw.JWT(jm))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authedRequest(t, jm, http.MethodDelete, "/api/user_info/avatar/", ""))
	return rec
}

// アバターを削除した時、200 が返ること。
func TestUserHandler_DeleteAvatar_Returns200(t *testing.T) {
	// Arrange & Act
	rec := serveDeleteAvatar(t, func(_ context.Context, _ string) (*domain.User, error) {
		return factory.NewUser(factory.WithID(testUserID)), nil
	})

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}
