package repository

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil"
)

func TestUserRepo_CreateAndFind(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)

	u := &domain.ApplicationUser{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "hashed",
		IsActive: true,
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == 0 {
		t.Fatalf("ID should be assigned after create")
	}

	byName, err := repo.FindByUsername(ctx, "alice")
	if err != nil || byName == nil {
		t.Fatalf("findByUsername: got=%v err=%v", byName, err)
	}
	if byName.Email != "alice@example.com" {
		t.Errorf("email = %q, want alice@example.com", byName.Email)
	}

	byEmail, err := repo.FindByEmail(ctx, "alice@example.com")
	if err != nil || byEmail == nil {
		t.Fatalf("findByEmail: got=%v err=%v", byEmail, err)
	}
	if byEmail.Username != "alice" {
		t.Errorf("username = %q, want alice", byEmail.Username)
	}

	byID, err := repo.FindByID(ctx, u.ID)
	if err != nil || byID == nil {
		t.Fatalf("findByID: got=%v err=%v", byID, err)
	}
	if byID.Username != "alice" {
		t.Errorf("username = %q, want alice", byID.Username)
	}
}

func TestUserRepo_Create_DuplicateUsername(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)

	first := &domain.ApplicationUser{Username: "alice", Email: "alice@example.com", Password: "x", IsActive: true}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// 同じ username（email は別）→ uniqueIndex 制約でエラーになること
	dup := &domain.ApplicationUser{Username: "alice", Email: "other@example.com", Password: "x", IsActive: true}
	if err := repo.Create(ctx, dup); err == nil {
		t.Fatalf("duplicate username should fail with unique constraint, got nil error")
	}
}

func TestUserRepo_Create_DuplicateEmail(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)

	first := &domain.ApplicationUser{Username: "alice", Email: "alice@example.com", Password: "x", IsActive: true}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// 同じ email（username は別）→ uniqueIndex 制約でエラーになること
	dup := &domain.ApplicationUser{Username: "bob", Email: "alice@example.com", Password: "x", IsActive: true}
	if err := repo.Create(ctx, dup); err == nil {
		t.Fatalf("duplicate email should fail with unique constraint, got nil error")
	}
}

func TestUserRepo_FindByUsername_NotFound(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)

	// 見つからない場合は (nil, nil)
	got, err := repo.FindByUsername(ctx, "nobody")
	if err != nil {
		t.Fatalf("findByUsername: unexpected err=%v", err)
	}
	if got != nil {
		t.Errorf("should be nil for missing user, got %+v", got)
	}
}

func TestUserRepo_FindAll_OrderedByID(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)

	seedUser(t, "alice")
	seedUser(t, "bob")
	seedUser(t, "carol")

	users, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("findAll: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("len = %d, want 3", len(users))
	}
	// ID 昇順で返ること
	if users[0].Username != "alice" || users[1].Username != "bob" || users[2].Username != "carol" {
		t.Errorf("order = [%s %s %s], want [alice bob carol]", users[0].Username, users[1].Username, users[2].Username)
	}
}
