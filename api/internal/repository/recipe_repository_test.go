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

// mustFindRecipe は FindByID で取得し、エラー・nil をガードして実体を返す。
func mustFindRecipe(t *testing.T, ctx context.Context, repo domain.RecipeRepository, id uint) *domain.Recipe {
	t.Helper()
	got, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got)
	return got
}

// arrangeRecipeWithRelations は owner/共有先/ラベル/食材/調味料を持つレシピを1件作成し、
// その ID を返す。Create 後の各関連の永続化を1テスト1観点で検証するための共通セットアップ。
func arrangeRecipeWithRelations(t *testing.T) (context.Context, domain.RecipeRepository, uint) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	r := factory.NewRecipe(
		factory.WithTitle("肉じゃが"),
		factory.WithOwnerID(owner.ID),
		factory.WithSharedUsers(*friend),
	)
	r.Servings = 2
	r.Labels = []domain.RecipeLabel{{Name: "和食"}}
	r.Ingredients = []domain.RecipeIngredient{{Name: "じゃがいも", Quantity: 3, Unit: "個"}}
	r.Seasonings = []domain.RecipeSeasoning{{Name: "醤油", Quantity: 2, Unit: "大さじ"}}
	require.NoError(t, repo.Create(ctx, r))
	return ctx, repo, r.ID
}

// 関連付きレシピを作成した時、FindByID で owner が読み込まれること。
func TestRecipeRepo_FindByID_LoadsOwner(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, "owner", got.Owner.Username)
}

// 関連付きレシピを作成した時、FindByID でラベルが構造体ごと読み込まれること。
func TestRecipeRepo_FindByID_LoadsLabels(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, []domain.RecipeLabel{{ID: 1, RecipeID: 1, Name: "和食"}}, got.Labels)
}

// 関連付きレシピを作成した時、FindByID で共有先ユーザーが読み込まれること。
func TestRecipeRepo_FindByID_LoadsSharedUsers(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	require.Len(t, got.SharedUsers, 1)
	assert.Equal(t, "friend", got.SharedUsers[0].Username)
}

// 関連付きレシピを作成した時、FindByID で食材が構造体ごと読み込まれること。
func TestRecipeRepo_FindByID_LoadsIngredients(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, []domain.RecipeIngredient{{
		ID:       1,
		RecipeID: 1,
		Name:     "じゃがいも",
		Quantity: 3,
		Unit:     "個",
	}}, got.Ingredients)
}

// 関連付きレシピを作成した時、FindByID で調味料が構造体ごと読み込まれること。
func TestRecipeRepo_FindByID_LoadsSeasonings(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, []domain.RecipeSeasoning{{
		ID:       1,
		RecipeID: 1,
		Name:     "醤油",
		Quantity: 2,
		Unit:     "大さじ",
	}}, got.Seasonings)
}

// arrangeSharedRecipe は owner が持ち friend に共有したレシピを1件作成し、
// owner / friend / stranger の3ユーザーを返す。
func arrangeSharedRecipe(t *testing.T) (context.Context, domain.RecipeRepository, *domain.User, *domain.User, *domain.User) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	stranger := seedUser(t, "stranger")
	r := factory.NewRecipe(
		factory.WithTitle("共有レシピ"),
		factory.WithOwnerID(owner.ID),
		factory.WithSharedUsers(*friend),
	)
	require.NoError(t, repo.Create(ctx, r))
	return ctx, repo, owner, friend, stranger
}

// 共有レシピがある時、owner の FindAllForUser でそのレシピが返ること。
func TestRecipeRepo_FindAllForUser_OwnerSees(t *testing.T) {
	// Arrange
	ctx, repo, owner, _, _ := arrangeSharedRecipe(t)

	// Act
	list, err := repo.FindAllForUser(ctx, owner.ID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

// 共有レシピがある時、共有先の FindAllForUser でそのレシピが返ること。
func TestRecipeRepo_FindAllForUser_SharedUserSees(t *testing.T) {
	// Arrange
	ctx, repo, _, friend, _ := arrangeSharedRecipe(t)

	// Act
	list, err := repo.FindAllForUser(ctx, friend.ID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

// 共有レシピがある時、無関係なユーザーの FindAllForUser では空が返ること。
func TestRecipeRepo_FindAllForUser_StrangerDoesNotSee(t *testing.T) {
	// Arrange
	ctx, repo, _, _, stranger := arrangeSharedRecipe(t)

	// Act
	list, err := repo.FindAllForUser(ctx, stranger.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, list)
}

// arrangeUpdatedRecipe は初版レシピを作成後、別の食材・別ラベルに差し替える Update を実行し、
// その ID を返す。置き換えセマンティクスを1テスト1観点で検証するための共通セットアップ。
func arrangeUpdatedRecipe(t *testing.T) (context.Context, domain.RecipeRepository, uint) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	r := factory.NewRecipe(factory.WithTitle("初版"), factory.WithOwnerID(owner.ID))
	r.Labels = []domain.RecipeLabel{{Name: "和食"}}
	r.Ingredients = []domain.RecipeIngredient{{Name: "じゃがいも", Quantity: 1, Unit: "個"}}
	require.NoError(t, repo.Create(ctx, r))

	r.Title = "改訂版"
	r.Labels = []domain.RecipeLabel{{Name: "夕食"}}
	r.Ingredients = []domain.RecipeIngredient{{Name: "人参", Quantity: 2, Unit: "本"}}
	r.Seasonings = nil
	require.NoError(t, repo.Update(ctx, r))
	return ctx, repo, r.ID
}

// レシピを更新した時、タイトルが置き換わること。
func TestRecipeRepo_Update_ReplacesTitle(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeUpdatedRecipe(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, "改訂版", got.Title)
}

// レシピを更新した時、食材が行ごと新しいものに置き換わること。
func TestRecipeRepo_Update_ReplacesIngredients(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeUpdatedRecipe(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert: 旧行(id1)は削除され、新規行(id2)に置き換わる。
	assert.Equal(t, []domain.RecipeIngredient{{
		ID:       2,
		RecipeID: 1,
		Name:     "人参",
		Quantity: 2,
		Unit:     "本",
	}}, got.Ingredients)
}

// レシピを更新した時、ラベルが行ごと新しいものに置き換わること。
func TestRecipeRepo_Update_ReplacesLabels(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeUpdatedRecipe(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, []domain.RecipeLabel{{ID: 2, RecipeID: 1, Name: "夕食"}}, got.Labels)
}

// arrangeDeletedRecipe は子テーブル(食材)と共有先を持つレシピを作成後 Delete を実行し、その ID を返す。
func arrangeDeletedRecipe(t *testing.T) (context.Context, domain.RecipeRepository, uint) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	r := factory.NewRecipe(
		factory.WithTitle("削除対象"),
		factory.WithOwnerID(owner.ID),
		factory.WithSharedUsers(*friend),
	)
	r.Ingredients = []domain.RecipeIngredient{{Name: "卵", Quantity: 2, Unit: "個"}}
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.Delete(ctx, r))
	return ctx, repo, r.ID
}

// レシピを削除した時、本体が消えること。
func TestRecipeRepo_Delete_RemovesRecipe(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeDeletedRecipe(t)

	// Act
	got, err := repo.FindByID(ctx, id)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, got) // 本体が消えること
}

// レシピを削除した時、子テーブル(recipe_ingredients)も FK CASCADE で消えること。
func TestRecipeRepo_Delete_CascadesIngredientRows(t *testing.T) {
	// Arrange
	_, _, id := arrangeDeletedRecipe(t)

	// Act
	var count int64
	testDB.Model(&domain.RecipeIngredient{}).Where("recipe_id = ?", id).Count(&count)

	// Assert
	assert.Zero(t, count)
}

// レシピを削除した時、中間テーブル(recipe_shares)も FK CASCADE で消えること。
func TestRecipeRepo_Delete_CascadesShareRows(t *testing.T) {
	// Arrange
	_, _, id := arrangeDeletedRecipe(t)

	// Act
	var count int64
	testDB.Table("recipe_shares").Where("recipe_id = ?", id).Count(&count)

	// Assert
	assert.Zero(t, count)
}
