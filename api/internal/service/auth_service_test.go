package service

import (
	"context"
	"errors"
	"testing"

	"recipe-backend/internal/domain"
	jwtpkg "recipe-backend/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

func newAuthService(users map[string]*domain.ApplicationUser) (AuthService, *jwtpkg.Manager) {
	ur := &mockUserRepo{byName: users}
	jm := jwtpkg.NewManager("test-secret")
	return NewAuthService(ur, jm), jm
}

func hashed(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return string(h)
}

func TestAuthLogin_Success(t *testing.T) {
	svc, jm := newAuthService(map[string]*domain.ApplicationUser{
		"alice": {ID: 1, Username: "alice", Password: hashed(t, "pw"), IsActive: true},
	})

	access, refresh, err := svc.Login(context.Background(), "alice", "pw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("expected non-empty tokens")
	}
	uid, err := jm.Parse(access, jwtpkg.TypeAccess)
	if err != nil || uid != 1 {
		t.Errorf("access token uid = %d, err = %v; want uid 1", uid, err)
	}
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	svc, _ := newAuthService(map[string]*domain.ApplicationUser{
		"alice": {ID: 1, Username: "alice", Password: hashed(t, "pw"), IsActive: true},
	})
	_, _, err := svc.Login(context.Background(), "alice", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthLogin_NoSuchUser(t *testing.T) {
	svc, _ := newAuthService(map[string]*domain.ApplicationUser{})
	_, _, err := svc.Login(context.Background(), "ghost", "pw")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthLogin_InactiveUser(t *testing.T) {
	svc, _ := newAuthService(map[string]*domain.ApplicationUser{
		"alice": {ID: 1, Username: "alice", Password: hashed(t, "pw"), IsActive: false},
	})
	_, _, err := svc.Login(context.Background(), "alice", "pw")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthRefresh_Valid(t *testing.T) {
	svc, jm := newAuthService(map[string]*domain.ApplicationUser{})
	refresh, _ := jm.GenerateRefresh(5)

	access, err := svc.Refresh(context.Background(), refresh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	uid, err := jm.Parse(access, jwtpkg.TypeAccess)
	if err != nil || uid != 5 {
		t.Errorf("new access uid = %d, err = %v; want uid 5", uid, err)
	}
}

func TestAuthRefresh_RejectsAccessToken(t *testing.T) {
	svc, jm := newAuthService(map[string]*domain.ApplicationUser{})
	access, _ := jm.GenerateAccess(5) // access を refresh として渡す → 失敗

	if _, err := svc.Refresh(context.Background(), access); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthRefresh_Garbage(t *testing.T) {
	svc, _ := newAuthService(map[string]*domain.ApplicationUser{})
	if _, err := svc.Refresh(context.Background(), "bad-token"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthRegister_Success(t *testing.T) {
	ur := &mockUserRepo{}
	svc := NewAuthService(ur, jwtpkg.NewManager("test-secret"))

	user, err := svc.Register(context.Background(), "alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected an assigned ID")
	}
	if user.Password == "password123" {
		t.Error("password must be hashed, not stored as plaintext")
	}
	if stored, _ := ur.FindByUsername(context.Background(), "alice"); stored == nil {
		t.Error("user should be stored")
	}
}

func TestAuthRegister_DuplicateUsername(t *testing.T) {
	ur := &mockUserRepo{byName: map[string]*domain.ApplicationUser{
		"alice": {ID: 1, Username: "alice"},
	}}
	svc := NewAuthService(ur, jwtpkg.NewManager("s"))

	_, err := svc.Register(context.Background(), "alice", "new@example.com", "password123")
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("err = %v, want ErrUserAlreadyExists", err)
	}
}

func TestAuthRegister_DuplicateEmail(t *testing.T) {
	ur := &mockUserRepo{byEmail: map[string]*domain.ApplicationUser{
		"taken@example.com": {ID: 1, Email: "taken@example.com"},
	}}
	svc := NewAuthService(ur, jwtpkg.NewManager("s"))

	_, err := svc.Register(context.Background(), "newuser", "taken@example.com", "password123")
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("err = %v, want ErrUserAlreadyExists", err)
	}
}
