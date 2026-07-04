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

// seedRecipeWithLabels は owner のレシピをラベル付きで1件作成し、その ID を返す。
func seedRecipeWithLabels(t *testing.T, ctx context.Context, owner *domain.User, title string, labels ...string) string {
	t.Helper()
	repo := NewRecipeRepository(testDB)
	r := factory.NewRecipe(factory.WithTitle(title), factory.WithOwnerID(owner.ID))
	for _, name := range labels {
		r.Labels = append(r.Labels, domain.RecipeLabel{Name: name})
	}
	require.NoError(t, repo.Create(ctx, r))
	return r.ID
}

// recipeLabelNames は recipeID の recipe_labels の name を返す。
func recipeLabelNames(t *testing.T, recipeID string) []string {
	t.Helper()
	var names []string
	require.NoError(t, testDB.Model(&domain.RecipeLabel{}).
		Where("recipe_id = ?", recipeID).Pluck("name", &names).Error)
	return names
}

// ラベルを作成した時、FindAllForOwner で名前順に返ること。
func TestLabelRepo_Create_And_FindAllForOwner(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")

	// Act
	require.NoError(t, repo.Create(ctx, &domain.Label{Name: "洋食", OwnerID: owner.ID}))
	require.NoError(t, repo.Create(ctx, &domain.Label{Name: "和食", OwnerID: owner.ID}))

	// Assert: 名前昇順
	labels, err := repo.FindAllForOwner(ctx, owner.ID)
	require.NoError(t, err)
	names := make([]string, len(labels))
	for i := range labels {
		names[i] = labels[i].Name
	}
	assert.Equal(t, []string{"和食", "洋食"}, names)
}

// 他人のラベルは FindAllForOwner に含まれないこと。
func TestLabelRepo_FindAllForOwner_ExcludesOthers(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	other := seedUser(t, "other")
	require.NoError(t, repo.Create(ctx, &domain.Label{Name: "和食", OwnerID: other.ID}))

	// Act
	labels, err := repo.FindAllForOwner(ctx, owner.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, labels)
}

// 該当するラベルがある時、FindByOwnerAndName がそれを返すこと。
func TestLabelRepo_FindByOwnerAndName_Found(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	require.NoError(t, repo.Create(ctx, &domain.Label{Name: "和食", OwnerID: owner.ID}))

	// Act
	found, err := repo.FindByOwnerAndName(ctx, owner.ID, "和食")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "和食", found.Name)
}

// 該当するラベルが無い時、FindByOwnerAndName が nil を返すこと。
func TestLabelRepo_FindByOwnerAndName_Missing(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")

	// Act
	missing, err := repo.FindByOwnerAndName(ctx, owner.ID, "洋食")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, missing)
}

// ラベルを改名した時、所有者のレシピの recipe_labels にも伝播すること。
func TestLabelRepo_Rename_PropagatesToRecipeLabels(t *testing.T) {
	// Arrange: owner が「和食」ラベルと、それを付けたレシピを持つ
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	label := &domain.Label{Name: "和食", OwnerID: owner.ID}
	require.NoError(t, repo.Create(ctx, label))
	recipeID := seedRecipeWithLabels(t, ctx, owner, "肉じゃが", "和食")

	// Act
	require.NoError(t, repo.Rename(ctx, label, "日本料理"))

	// Assert: レシピ側のラベル名も変わる
	assert.Equal(t, []string{"日本料理"}, recipeLabelNames(t, recipeID))
}

// 改名先の名前を既に持つレシピでは、旧名を消して重複を避けること。
func TestLabelRepo_Rename_DedupesOnConflict(t *testing.T) {
	// Arrange: レシピに「和食」と「日本料理」の両方が付いている
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	label := &domain.Label{Name: "和食", OwnerID: owner.ID}
	require.NoError(t, repo.Create(ctx, label))
	recipeID := seedRecipeWithLabels(t, ctx, owner, "肉じゃが", "和食", "日本料理")

	// Act: 「和食」→「日本料理」(既に付いている)
	require.NoError(t, repo.Rename(ctx, label, "日本料理"))

	// Assert: 重複せず「日本料理」1件になる
	assert.Equal(t, []string{"日本料理"}, recipeLabelNames(t, recipeID))
}

// 改名は他人のレシピの同名ラベルには影響しないこと。
func TestLabelRepo_Rename_DoesNotTouchOthersRecipes(t *testing.T) {
	// Arrange: owner と other が別々に「和食」を使う
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	other := seedUser(t, "other")
	label := &domain.Label{Name: "和食", OwnerID: owner.ID}
	require.NoError(t, repo.Create(ctx, label))
	otherRecipe := seedRecipeWithLabels(t, ctx, other, "他人のレシピ", "和食")

	// Act
	require.NoError(t, repo.Rename(ctx, label, "日本料理"))

	// Assert: other のレシピの「和食」はそのまま
	assert.Equal(t, []string{"和食"}, recipeLabelNames(t, otherRecipe))
}

// ラベルを削除した時、マスタと所有者のレシピの recipe_labels から消えること。
func TestLabelRepo_Delete_RemovesFromMasterAndRecipes(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	owner := seedUser(t, "owner")
	label := &domain.Label{Name: "和食", OwnerID: owner.ID}
	require.NoError(t, repo.Create(ctx, label))
	recipeID := seedRecipeWithLabels(t, ctx, owner, "肉じゃが", "和食")

	// Act
	require.NoError(t, repo.Delete(ctx, label))

	// Assert: マスタから消え、レシピからも外れる
	got, err := repo.FindByID(ctx, label.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
	assert.Empty(t, recipeLabelNames(t, recipeID))
}
