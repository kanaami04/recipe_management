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
	label, _ := repo.GetOrCreateLabel(ctx, "和食")
	ing, _ := repo.GetOrCreateIngredient(ctx, "じゃがいも")
	sea, _ := repo.GetOrCreateSeasoning(ctx, "醤油")
	r := factory.NewRecipe(
		factory.WithTitle("肉じゃが"),
		factory.WithOwnerID(owner.ID),
		factory.WithSharedUsers(*friend),
	)
	r.CreateFor = 2
	r.Labels = []domain.RecipeLabel{*label}
	r.Cooking = []domain.Cooking{{IngredientID: ing.ID, Quantity: 3, Unit: "個"}}
	r.Season = []domain.Season{{SeasoningID: sea.ID, Quantity: 2, Unit: "大さじ"}}
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
	assert.Equal(t, []domain.RecipeLabel{{ID: 1, Name: "和食"}}, got.Labels)
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

// 関連付きレシピを作成した時、FindByID で食材(Cooking)が構造体ごと読み込まれること。
func TestRecipeRepo_FindByID_LoadsCooking(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, []domain.Cooking{{
		ID:           1,
		RecipeID:     1,
		IngredientID: 1,
		Ingredient:   domain.Ingredient{ID: 1, Name: "じゃがいも"},
		Quantity:     3,
		Unit:         "個",
	}}, got.Cooking)
}

// 関連付きレシピを作成した時、FindByID で調味料(Season)が構造体ごと読み込まれること。
func TestRecipeRepo_FindByID_LoadsSeason(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, []domain.Season{{
		ID:          1,
		RecipeID:    1,
		SeasoningID: 1,
		Seasoning:   domain.Seasoning{ID: 1, Name: "醤油"},
		Quantity:    2,
		Unit:        "大さじ",
	}}, got.Season)
}

// 同名のラベルを2回 GetOrCreate した時、既存行が再利用されること。
func TestRecipeRepo_GetOrCreateLabel_ReusesByName(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)

	// Act: 同じ name で2回 get-or-create する
	a, _ := repo.GetOrCreateLabel(ctx, "和食")
	b, _ := repo.GetOrCreateLabel(ctx, "和食")

	// Assert: 同名は既存行を再利用すること
	assert.Equal(t, a.ID, b.ID)
}

// 同名の食材を2回 GetOrCreate した時、既存行が再利用されること。
func TestRecipeRepo_GetOrCreateIngredient_ReusesByName(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)

	// Act: 同じ name で2回 get-or-create する
	i1, _ := repo.GetOrCreateIngredient(ctx, "塩")
	i2, _ := repo.GetOrCreateIngredient(ctx, "塩")

	// Assert: 同名は既存行を再利用すること
	assert.Equal(t, i1.ID, i2.ID)
}

// arrangeSharedRecipe は owner が持ち friend に共有したレシピを1件作成し、
// owner / friend / stranger の3ユーザーを返す。
func arrangeSharedRecipe(t *testing.T) (context.Context, domain.RecipeRepository, *domain.ApplicationUser, *domain.ApplicationUser, *domain.ApplicationUser) {
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
	ingA, _ := repo.GetOrCreateIngredient(ctx, "じゃがいも")
	labelA, _ := repo.GetOrCreateLabel(ctx, "和食")
	r := factory.NewRecipe(factory.WithTitle("初版"), factory.WithOwnerID(owner.ID))
	r.Labels = []domain.RecipeLabel{*labelA}
	r.Cooking = []domain.Cooking{{IngredientID: ingA.ID, Quantity: 1, Unit: "個"}}
	require.NoError(t, repo.Create(ctx, r))

	ingB, _ := repo.GetOrCreateIngredient(ctx, "人参")
	labelB, _ := repo.GetOrCreateLabel(ctx, "夕食")
	r.Title = "改訂版"
	r.Labels = []domain.RecipeLabel{*labelB}
	r.Cooking = []domain.Cooking{{IngredientID: ingB.ID, Quantity: 2, Unit: "本"}}
	r.Season = nil
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

// レシピを更新した時、食材(Cooking)が構造体ごと新しいものに置き換わること。
func TestRecipeRepo_Update_ReplacesCooking(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeUpdatedRecipe(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert: 旧 cooking(id1) は削除され、新規行(id2)に置き換わる。
	assert.Equal(t, []domain.Cooking{{
		ID:           2,
		RecipeID:     1,
		IngredientID: 2,
		Ingredient:   domain.Ingredient{ID: 2, Name: "人参"},
		Quantity:     2,
		Unit:         "本",
	}}, got.Cooking)
}

// レシピを更新した時、ラベルが構造体ごと新しいものに置き換わること。
func TestRecipeRepo_Update_ReplacesLabels(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeUpdatedRecipe(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, []domain.RecipeLabel{{ID: 2, Name: "夕食"}}, got.Labels)
}

// arrangeDeletedRecipe は子テーブル(cooking)を持つレシピを作成後 Delete を実行し、その ID を返す。
func arrangeDeletedRecipe(t *testing.T) (context.Context, domain.RecipeRepository, uint) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	ing, _ := repo.GetOrCreateIngredient(ctx, "卵")
	r := factory.NewRecipe(factory.WithTitle("削除対象"), factory.WithOwnerID(owner.ID))
	r.Cooking = []domain.Cooking{{IngredientID: ing.ID, Quantity: 2, Unit: "個"}}
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

// レシピを削除した時、子テーブル(cooking)も消えること。
func TestRecipeRepo_Delete_RemovesCookingRows(t *testing.T) {
	// Arrange
	_, _, id := arrangeDeletedRecipe(t)

	// Act
	var cookingCount int64
	testDB.Model(&domain.Cooking{}).Where("recipe_id = ?", id).Count(&cookingCount)

	// Assert
	assert.Zero(t, cookingCount) // 子テーブル(cooking)も消えること
}
