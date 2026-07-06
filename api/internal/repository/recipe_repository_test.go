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
func mustFindRecipe(t *testing.T, ctx context.Context, repo domain.RecipeRepository, id string) *domain.Recipe {
	t.Helper()
	got, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got)
	return got
}

// arrangeRecipeWithRelations は owner/共有先/ラベル/食材/調味料を持つレシピを1件作成し、
// その ID を返す。Create 後の各関連の永続化を1テスト1観点で検証するための共通セットアップ。
func arrangeRecipeWithRelations(t *testing.T) (context.Context, domain.RecipeRepository, string) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	r := factory.NewRecipe(
		factory.WithTitle("肉じゃが"),
		factory.WithOwnerID(owner.ID),
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
	require.Len(t, got.Labels, 1)
	assert.Equal(t, "和食", got.Labels[0].Name)
	assert.Equal(t, got.ID, got.Labels[0].RecipeID)
}

// 関連付きレシピを作成した時、FindByID で食材が構造体ごと読み込まれること。
func TestRecipeRepo_FindByID_LoadsIngredients(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	require.Len(t, got.Ingredients, 1)
	assert.Equal(t, got.ID, got.Ingredients[0].RecipeID)
	assert.Equal(t, "じゃがいも", got.Ingredients[0].Name)
	assert.Equal(t, 3.0, got.Ingredients[0].Quantity)
	assert.Equal(t, "個", got.Ingredients[0].Unit)
}

// 関連付きレシピを作成した時、FindByID で調味料が構造体ごと読み込まれること。
func TestRecipeRepo_FindByID_LoadsSeasonings(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	require.Len(t, got.Seasonings, 1)
	assert.Equal(t, got.ID, got.Seasonings[0].RecipeID)
	assert.Equal(t, "醤油", got.Seasonings[0].Name)
	assert.Equal(t, 2.0, got.Seasonings[0].Quantity)
	assert.Equal(t, "大さじ", got.Seasonings[0].Unit)
}

// arrangeSharedRecipe は owner が持つレシピを1件作成し、owner と friend を同じシェアグループに
// 入れる。owner / friend / stranger の3ユーザーを返す(friend は共有で見え、stranger は見えない)。
func arrangeSharedRecipe(t *testing.T) (context.Context, domain.RecipeRepository, *domain.User, *domain.User, *domain.User) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	stranger := seedUser(t, "stranger")
	seedShareGroup(t, owner, friend)
	r := factory.NewRecipe(
		factory.WithTitle("共有レシピ"),
		factory.WithOwnerID(owner.ID),
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

// 共有レシピがある時、同じグループのメンバーの FindAllForUser でそのレシピが返ること。
func TestRecipeRepo_FindAllForUser_GroupMemberSees(t *testing.T) {
	// Arrange
	ctx, repo, _, friend, _ := arrangeSharedRecipe(t)

	// Act
	list, err := repo.FindAllForUser(ctx, friend.ID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

// 共有レシピがある時、グループ外ユーザーの FindAllForUser では空が返ること。
func TestRecipeRepo_FindAllForUser_StrangerDoesNotSee(t *testing.T) {
	// Arrange
	ctx, repo, _, _, stranger := arrangeSharedRecipe(t)

	// Act
	list, err := repo.FindAllForUser(ctx, stranger.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, list)
}

// recipeIDs は取得したレシピのスライスから ID だけを取り出す。
func recipeIDs(recipes []domain.Recipe) []string {
	ids := make([]string, 0, len(recipes))
	for i := range recipes {
		ids = append(ids, recipes[i].ID)
	}
	return ids
}

// 並べ替えを保存した時、FindAllForUser がその順序で返すこと。
func TestRecipeRepo_Reorder_ReflectsInListOrder(t *testing.T) {
	// Arrange: owner のレシピを 3 件(既定は作成=id 昇順)
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	r1 := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	r2 := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	r3 := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r1))
	require.NoError(t, repo.Create(ctx, r2))
	require.NoError(t, repo.Create(ctx, r3))

	// Act: 逆順に並べ替える
	want := []string{r3.ID, r1.ID, r2.ID}
	require.NoError(t, repo.Reorder(ctx, owner.ID, want))

	// Assert
	list, err := repo.FindAllForUser(ctx, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, want, recipeIDs(list))
}

// あるユーザーが共有レシピを並べ替えても、別ユーザーの並び順には影響しないこと。
func TestRecipeRepo_Reorder_IsolatedPerUser(t *testing.T) {
	// Arrange: owner が 2 件を friend と共有(両者とも既定は id 昇順で見える)
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	seedShareGroup(t, owner, friend)
	a := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	b := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, a))
	require.NoError(t, repo.Create(ctx, b))

	// Act: owner だけ b, a の順に並べ替える
	require.NoError(t, repo.Reorder(ctx, owner.ID, []string{b.ID, a.ID}))

	// Assert: friend の並び順は既定(a, b)のままで影響を受けない
	friendList, err := repo.FindAllForUser(ctx, friend.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID, b.ID}, recipeIDs(friendList))
}

// archivedOf は list から id のレシピの Archived を返す(無ければ false)。
func archivedOf(list []domain.Recipe, id string) bool {
	for i := range list {
		if list[i].ID == id {
			return list[i].Archived
		}
	}
	return false
}

// アーカイブを保存した時、そのユーザーの FindAllForUser で Archived=true になること。
func TestRecipeRepo_SetArchived_ReflectsInList(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	r := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r))

	// Act
	require.NoError(t, repo.SetArchived(ctx, owner.ID, r.ID, true))

	// Assert
	list, err := repo.FindAllForUser(ctx, owner.ID)
	require.NoError(t, err)
	assert.True(t, archivedOf(list, r.ID))
}

// 共有相手がアーカイブしても、所有者の一覧では Archived のままにならないこと(ユーザーごと)。
func TestRecipeRepo_SetArchived_IsolatedPerUser(t *testing.T) {
	// Arrange: owner が friend に共有したレシピ
	ctx, repo, owner, friend, _ := arrangeSharedRecipe(t)
	list, err := repo.FindAllForUser(ctx, friend.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	recipeID := list[0].ID

	// Act: friend が自分の状態としてアーカイブする
	require.NoError(t, repo.SetArchived(ctx, friend.ID, recipeID, true))

	// Assert: friend では Archived、owner では非 Archived
	friendList, err := repo.FindAllForUser(ctx, friend.ID)
	require.NoError(t, err)
	assert.True(t, archivedOf(friendList, recipeID))
	ownerList, err := repo.FindAllForUser(ctx, owner.ID)
	require.NoError(t, err)
	assert.False(t, archivedOf(ownerList, recipeID))
}

// アーカイブを解除した時、そのユーザーの FindAllForUser で Archived=false に戻ること。
func TestRecipeRepo_SetArchived_Unarchive(t *testing.T) {
	// Arrange: 一度アーカイブする
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	r := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.SetArchived(ctx, owner.ID, r.ID, true))

	// Act
	require.NoError(t, repo.SetArchived(ctx, owner.ID, r.ID, false))

	// Assert
	list, err := repo.FindAllForUser(ctx, owner.ID)
	require.NoError(t, err)
	assert.False(t, archivedOf(list, r.ID))
}

// IsArchived は、アーカイブ済みのユーザーには true、そうでないユーザーには false を返すこと。
func TestRecipeRepo_IsArchived_PerUser(t *testing.T) {
	// Arrange: owner が r をアーカイブ、friend はしていない
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	r := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.SetArchived(ctx, owner.ID, r.ID, true))

	// Act
	ownerArchived, err := repo.IsArchived(ctx, owner.ID, r.ID)
	require.NoError(t, err)
	friendArchived, err := repo.IsArchived(ctx, friend.ID, r.ID)
	require.NoError(t, err)

	// Assert
	assert.True(t, ownerArchived)
	assert.False(t, friendArchived)
}

// レシピを削除した時、その recipe_archives 行も FK CASCADE で消えること。
func TestRecipeRepo_Delete_CascadesRecipeArchives(t *testing.T) {
	// Arrange: アーカイブ行を作ってから削除する
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	r := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.SetArchived(ctx, owner.ID, r.ID, true))

	// Act
	require.NoError(t, repo.Delete(ctx, r))

	// Assert
	var count int64
	testDB.Table("recipe_archives").Where("recipe_id = ?", r.ID).Count(&count)
	assert.Zero(t, count)
}

// レシピを削除した時、その recipe_orders 行も FK CASCADE で消えること。
func TestRecipeRepo_Delete_CascadesRecipeOrders(t *testing.T) {
	// Arrange: 並べ替えで recipe_orders 行を作ってから削除する
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	r := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.Reorder(ctx, owner.ID, []string{r.ID}))

	// Act
	require.NoError(t, repo.Delete(ctx, r))

	// Assert
	var count int64
	testDB.Table("recipe_orders").Where("recipe_id = ?", r.ID).Count(&count)
	assert.Zero(t, count)
}

// countRows は table の (user_id, recipe_id) に一致する行数を返す。掃除結果の検証に使う。
func countRows(t *testing.T, table, userID, recipeID string) int64 {
	t.Helper()
	var count int64
	if err := testDB.Table(table).
		Where("user_id = ? AND recipe_id = ?", userID, recipeID).
		Count(&count).Error; err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}

// 共有が解けて見えなくなったレシピについて、PruneRecipeState がそのユーザーの
// recipe_archives 行を消すこと。
func TestRecipeRepo_PruneRecipeState_RemovesOrphanedArchive(t *testing.T) {
	// Arrange: friend が owner のレシピをアーカイブした後、グループから外れる
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	group := seedShareGroup(t, owner, friend)
	r := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.SetArchived(ctx, friend.ID, r.ID, true))
	require.NoError(t, NewShareGroupRepository(testDB).RemoveMember(ctx, group.ID, friend.ID))

	// Act
	require.NoError(t, repo.PruneRecipeState(ctx, friend.ID))

	// Assert
	assert.Zero(t, countRows(t, "recipe_archives", friend.ID, r.ID))
}

// 共有が解けて見えなくなったレシピについて、PruneRecipeState がそのユーザーの
// recipe_orders 行を消すこと。
func TestRecipeRepo_PruneRecipeState_RemovesOrphanedOrder(t *testing.T) {
	// Arrange: friend が owner のレシピを並べ替えた後、グループから外れる
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	group := seedShareGroup(t, owner, friend)
	r := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.Reorder(ctx, friend.ID, []string{r.ID}))
	require.NoError(t, NewShareGroupRepository(testDB).RemoveMember(ctx, group.ID, friend.ID))

	// Act
	require.NoError(t, repo.PruneRecipeState(ctx, friend.ID))

	// Assert
	assert.Zero(t, countRows(t, "recipe_orders", friend.ID, r.ID))
}

// 自分が所有するレシピについては、PruneRecipeState がアーカイブ行を残すこと(見えるため)。
func TestRecipeRepo_PruneRecipeState_KeepsOwnRecipe(t *testing.T) {
	// Arrange: friend が自分のレシピをアーカイブし、グループから外れる
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	group := seedShareGroup(t, owner, friend)
	r := factory.NewRecipe(factory.WithOwnerID(friend.ID))
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.SetArchived(ctx, friend.ID, r.ID, true))
	require.NoError(t, NewShareGroupRepository(testDB).RemoveMember(ctx, group.ID, friend.ID))

	// Act
	require.NoError(t, repo.PruneRecipeState(ctx, friend.ID))

	// Assert
	assert.Equal(t, int64(1), countRows(t, "recipe_archives", friend.ID, r.ID))
}

// まだ同じグループで見えるレシピについては、PruneRecipeState がアーカイブ行を残すこと。
func TestRecipeRepo_PruneRecipeState_KeepsStillSharedRecipe(t *testing.T) {
	// Arrange: friend が owner のレシピをアーカイブし、グループには留まったまま
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	seedShareGroup(t, owner, friend)
	r := factory.NewRecipe(factory.WithOwnerID(owner.ID))
	require.NoError(t, repo.Create(ctx, r))
	require.NoError(t, repo.SetArchived(ctx, friend.ID, r.ID, true))

	// Act
	require.NoError(t, repo.PruneRecipeState(ctx, friend.ID))

	// Assert
	assert.Equal(t, int64(1), countRows(t, "recipe_archives", friend.ID, r.ID))
}

// arrangeUpdatedRecipe は初版レシピを作成後、別の食材・別ラベルに差し替える Update を実行し、
// その ID を返す。置き換えセマンティクスを1テスト1観点で検証するための共通セットアップ。
func arrangeUpdatedRecipe(t *testing.T) (context.Context, domain.RecipeRepository, string) {
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

	// Assert: 旧行は削除され、新規行に置き換わる。
	require.Len(t, got.Ingredients, 1)
	assert.Equal(t, "人参", got.Ingredients[0].Name)
	assert.Equal(t, 2.0, got.Ingredients[0].Quantity)
	assert.Equal(t, "本", got.Ingredients[0].Unit)
}

// レシピを更新した時、ラベルが行ごと新しいものに置き換わること。
func TestRecipeRepo_Update_ReplacesLabels(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeUpdatedRecipe(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	require.Len(t, got.Labels, 1)
	assert.Equal(t, "夕食", got.Labels[0].Name)
}

// arrangeDeletedRecipe は子テーブル(食材)を持つレシピを作成後 Delete を実行し、その ID を返す。
func arrangeDeletedRecipe(t *testing.T) (context.Context, domain.RecipeRepository, string) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)
	owner := seedUser(t, "owner")
	r := factory.NewRecipe(
		factory.WithTitle("削除対象"),
		factory.WithOwnerID(owner.ID),
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
