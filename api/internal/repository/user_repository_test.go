package repository

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// arrangeUserRepo は結合テストの前提を整え、alice を1人作成済みの repo を返す。
// Create / Find 系の検証を「1テスト1観点」に分割するための共通セットアップ。
func arrangeUserRepo(t *testing.T) (context.Context, domain.UserRepository, *domain.ApplicationUser) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)
	alice := factory.NewUser(factory.WithUsername("alice"), factory.WithEmail("alice@example.com"))
	require.NoError(t, repo.Create(ctx, alice))
	return ctx, repo, alice
}

// ユーザーを作成した時、ID が採番されること。
func TestUserRepo_Create_AssignsID(t *testing.T) {
	// Arrange & Act: 作成はヘルパー内で実行される
	_, _, alice := arrangeUserRepo(t)

	// Assert
	assert.NotZero(t, alice.ID) // 作成後にIDが採番されること
}

// 作成済みユーザーがいる時、FindByUsername で該当ユーザーが返ること。
func TestUserRepo_FindByUsername_ReturnsUser(t *testing.T) {
	// Arrange
	ctx, repo, _ := arrangeUserRepo(t)

	// Act
	got, err := repo.FindByUsername(ctx, "alice")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "alice@example.com", got.Email)
}

// 作成済みユーザーがいる時、FindByEmail で該当ユーザーが返ること。
func TestUserRepo_FindByEmail_ReturnsUser(t *testing.T) {
	// Arrange
	ctx, repo, _ := arrangeUserRepo(t)

	// Act
	got, err := repo.FindByEmail(ctx, "alice@example.com")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "alice", got.Username)
}

// 作成済みユーザーがいる時、FindByID で該当ユーザーが返ること。
func TestUserRepo_FindByID_ReturnsUser(t *testing.T) {
	// Arrange
	ctx, repo, alice := arrangeUserRepo(t)

	// Act
	got, err := repo.FindByID(ctx, alice.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "alice", got.Username)
}

// username が既存と重複する時、Create がエラーになること。
func TestUserRepo_Create_DuplicateUsername(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)
	require.NoError(t, repo.Create(ctx, factory.NewUser(factory.WithUsername("alice"), factory.WithEmail("alice@example.com"))))

	// Act: 同じ username（email は別）で再作成
	err := repo.Create(ctx, factory.NewUser(factory.WithUsername("alice"), factory.WithEmail("other@example.com")))

	// Assert: uniqueIndex 制約でエラーになること
	assert.Error(t, err)
}

// email が既存と重複する時、Create がエラーになること。
func TestUserRepo_Create_DuplicateEmail(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)
	require.NoError(t, repo.Create(ctx, factory.NewUser(factory.WithUsername("alice"), factory.WithEmail("alice@example.com"))))

	// Act: 同じ email（username は別）で再作成
	err := repo.Create(ctx, factory.NewUser(factory.WithUsername("bob"), factory.WithEmail("alice@example.com")))

	// Assert: uniqueIndex 制約でエラーになること
	assert.Error(t, err)
}

// 該当ユーザーがいない時、FindByUsername で (nil, nil) が返ること。
func TestUserRepo_FindByUsername_NotFound(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)

	// Act
	got, err := repo.FindByUsername(ctx, "nobody")

	// Assert: 見つからない場合は (nil, nil)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// 複数ユーザーがいる時、FindAll で ID 昇順のまま全件が返ること。
func TestUserRepo_FindAll_OrderedByID(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewUserRepository(testDB)
	seedUser(t, "alice")
	seedUser(t, "bob")
	seedUser(t, "carol")

	// Act
	users, err := repo.FindAll(ctx)

	// Assert: ID 昇順で全件返ること
	require.NoError(t, err)
	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Username
	}
	assert.Equal(t, []string{"alice", "bob", "carol"}, names)
}
