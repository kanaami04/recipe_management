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
func arrangeUserRepo(t *testing.T) (context.Context, domain.UserRepository, *domain.User) {
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

// プロフィールを更新した時、username が保存されること。
func TestUserRepo_Update_SavesUsername(t *testing.T) {
	// Arrange
	ctx, repo, alice := arrangeUserRepo(t)
	alice.Username = "alice2"
	alice.Email = "alice2@example.com"

	// Act
	require.NoError(t, repo.Update(ctx, alice))

	// Assert
	got, err := repo.FindByID(ctx, alice.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "alice2", got.Username)
}

// プロフィールを更新した時、email が保存されること。
func TestUserRepo_Update_SavesEmail(t *testing.T) {
	// Arrange
	ctx, repo, alice := arrangeUserRepo(t)
	alice.Username = "alice2"
	alice.Email = "alice2@example.com"

	// Act
	require.NoError(t, repo.Update(ctx, alice))

	// Assert
	got, err := repo.FindByID(ctx, alice.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "alice2@example.com", got.Email)
}

// パスワードを更新した時、password_hash が保存されること。
func TestUserRepo_UpdatePassword_SavesHash(t *testing.T) {
	// Arrange
	ctx, repo, alice := arrangeUserRepo(t)

	// Act
	require.NoError(t, repo.UpdatePassword(ctx, alice.ID, "new-hash"))

	// Assert
	got, err := repo.FindByID(ctx, alice.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "new-hash", got.PasswordHash)
}

// アバターキーを設定した時、avatar_key が保存されること。
func TestUserRepo_UpdateAvatarKey_SavesKey(t *testing.T) {
	// Arrange
	ctx, repo, alice := arrangeUserRepo(t)
	key := "avatars/" + alice.ID + "/abc"

	// Act
	require.NoError(t, repo.UpdateAvatarKey(ctx, alice.ID, &key))

	// Assert
	got, err := repo.FindByID(ctx, alice.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.AvatarKey)
	assert.Equal(t, key, *got.AvatarKey)
}

// アバターキーを nil に戻した時、avatar_key が消えること。
func TestUserRepo_UpdateAvatarKey_ClearsKey(t *testing.T) {
	// Arrange: 一度設定してから
	ctx, repo, alice := arrangeUserRepo(t)
	key := "avatars/" + alice.ID + "/abc"
	require.NoError(t, repo.UpdateAvatarKey(ctx, alice.ID, &key))

	// Act
	require.NoError(t, repo.UpdateAvatarKey(ctx, alice.ID, nil))

	// Assert
	got, err := repo.FindByID(ctx, alice.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.AvatarKey)
}

// アカウントを削除した時、ユーザーが消えること。
func TestUserRepo_Delete_RemovesUser(t *testing.T) {
	// Arrange
	ctx, repo, alice := arrangeUserRepo(t)

	// Act
	require.NoError(t, repo.Delete(ctx, alice.ID))

	// Assert
	got, err := repo.FindByID(ctx, alice.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// アカウントを削除した時、所有レシピも消えること(owner FK は CASCADE 無しのため明示削除)。
func TestUserRepo_Delete_CascadesOwnedRecipes(t *testing.T) {
	// Arrange: alice が1件レシピを所有
	ctx, repo, alice := arrangeUserRepo(t)
	recipeRepo := NewRecipeRepository(testDB)
	r := factory.NewRecipe(factory.WithOwnerID(alice.ID))
	require.NoError(t, recipeRepo.Create(ctx, r))

	// Act
	require.NoError(t, repo.Delete(ctx, alice.ID))

	// Assert
	var count int64
	testDB.Model(&domain.Recipe{}).Where("owner_id = ?", alice.ID).Count(&count)
	assert.Zero(t, count)
}

// アカウントを削除した時、自分のラベルも FK CASCADE で消えること。
func TestUserRepo_Delete_CascadesOwnedLabels(t *testing.T) {
	// Arrange
	ctx, repo, alice := arrangeUserRepo(t)
	require.NoError(t, NewLabelRepository(testDB).Create(ctx, &domain.Label{Name: "和食", OwnerID: alice.ID}))

	// Act
	require.NoError(t, repo.Delete(ctx, alice.ID))

	// Assert
	var count int64
	testDB.Model(&domain.Label{}).Where("owner_id = ?", alice.ID).Count(&count)
	assert.Zero(t, count)
}

// アカウントを削除しても、同じグループの他人が所有していたレシピは残ること。
func TestUserRepo_Delete_KeepsOthersSharedRecipe(t *testing.T) {
	// Arrange: bob 所有のレシピを、bob と alice が同じグループで共有
	ctx, repo, alice := arrangeUserRepo(t)
	bob := seedUser(t, "bob")
	seedShareGroup(t, bob, alice) // 所有者 bob、メンバー alice
	recipeRepo := NewRecipeRepository(testDB)
	shared := factory.NewRecipe(factory.WithOwnerID(bob.ID))
	require.NoError(t, recipeRepo.Create(ctx, shared))

	// Act: alice を削除
	require.NoError(t, repo.Delete(ctx, alice.ID))

	// Assert: bob のレシピは残る
	got, err := recipeRepo.FindByID(ctx, shared.ID)
	require.NoError(t, err)
	assert.NotNil(t, got)
}
