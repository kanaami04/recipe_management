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

// arrangeLabelRepo は結合テストの前提を整え、空の LabelRepository を返す。
func arrangeLabelRepo(t *testing.T) (context.Context, domain.LabelRepository) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	return context.Background(), NewLabelRepository(testDB)
}

// seedRecipeWithLabels は owner のレシピをラベル付きで1件作成する。
func seedRecipeWithLabels(t *testing.T, ctx context.Context, owner *domain.User, title string, labels []string, shared ...domain.User) {
	t.Helper()
	repo := NewRecipeRepository(testDB)
	r := factory.NewRecipe(factory.WithTitle(title), factory.WithOwnerID(owner.ID), factory.WithSharedUsers(shared...))
	for _, name := range labels {
		r.Labels = append(r.Labels, domain.RecipeLabel{Name: name})
	}
	require.NoError(t, repo.Create(ctx, r))
}

// 自分のレシピに同名ラベルが複数付いている時、重複なく全ラベル名が返ること。
func TestLabelRepo_FindNamesForUser_DistinctNames(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	seedRecipeWithLabels(t, ctx, owner, "肉じゃが", []string{"和食", "夕食"})
	seedRecipeWithLabels(t, ctx, owner, "味噌汁", []string{"和食"})

	// Act
	got, err := repo.FindNamesForUser(ctx, owner.ID)

	// Assert: 同名ラベルは1つにまとまること
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"和食", "夕食"}, got)
}

// レシピを共有されている時、共有先ユーザーにもそのラベル名が返ること。
func TestLabelRepo_FindNamesForUser_IncludesSharedRecipes(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	seedRecipeWithLabels(t, ctx, owner, "共有レシピ", []string{"和食"}, *friend)

	// Act
	got, err := repo.FindNamesForUser(ctx, friend.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"和食"}, got)
}

// 他人のレシピのラベルは返らないこと。
func TestLabelRepo_FindNamesForUser_ExcludesOthers(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	stranger := seedUser(t, "stranger")
	seedRecipeWithLabels(t, ctx, owner, "非公開レシピ", []string{"和食"})

	// Act
	got, err := repo.FindNamesForUser(ctx, stranger.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, got)
}

// ラベルが1件もない時、空が返ること。
func TestLabelRepo_FindNamesForUser_Empty(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")

	// Act
	got, err := repo.FindNamesForUser(ctx, owner.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, got)
}
