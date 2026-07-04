package service

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// レシピを作成した時、owner・デフォルト値・各関連まで反映されたレシピが構築されること。
func TestRecipeCreate_BuildsRecipe(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	svc := NewRecipeService(rr, &mockUserRepo{})
	req := request.RecipeRequest{
		Title: "カレー",
		// create_for 未指定 → 1 に正規化される想定
		Label:   []request.LabelInput{{Name: "夕食"}},
		Cooking: []request.CookingInput{{Ingredients: request.NameInput{Name: "玉ねぎ"}, Quantity: 2, Unit: "個"}},
		Season:  []request.SeasonInput{{Seasoning: request.NameInput{Name: "塩"}, Quantity: 1, Unit: "g"}},
	}

	// Act
	_, err := svc.Create(context.Background(), "u42", req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, rr.created)
	want := domain.Recipe{
		ID:          rr.created.ID, // 採番された UUID をそのまま採用
		Title:       "カレー",
		Servings:    1, // 未指定はデフォルト1
		OwnerID:     "u42",
		Labels:      []domain.RecipeLabel{{Name: "夕食"}},
		SharedUsers: []domain.User{},
		Ingredients: []domain.RecipeIngredient{{Name: "玉ねぎ", Quantity: 2, Unit: "個"}},
		Seasonings:  []domain.RecipeSeasoning{{Name: "塩", Quantity: 1, Unit: "g"}},
	}
	assert.Equal(t, want, *rr.created)
}

// 存在しないユーザーを共有先に指定した時、ErrSharedUserNotFound が返ること。
func TestRecipeCreate_SharedUserNotFound(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	svc := NewRecipeService(rr, &mockUserRepo{})
	req := request.RecipeRequest{
		Title:      "親子丼",
		SharedUser: []request.SharedUserInput{{Username: "ghost"}},
	}

	// Act
	_, err := svc.Create(context.Background(), "u1", req)

	// Assert
	assert.ErrorIs(t, err, ErrSharedUserNotFound)
}

// 自分のレシピがある時、List でそのレシピが返ること。
func TestRecipeList_ReturnsRecipes(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u5"))
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	recipes, err := svc.List(context.Background(), "u5")

	// Assert
	require.NoError(t, err)
	assert.Len(t, recipes, 1)
}

// owner でも共有先でもないユーザーが更新する時、ErrForbidden が返ること。
func TestRecipeUpdate_ForbiddenForNonOwnerNonShared(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u100")) // 所有者は u100
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	_, err := svc.Update(context.Background(), "u999", "r1", request.RecipeRequest{Title: "x"}) // 別人 u999

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 共有先ユーザーが更新した時、タイトルが変わり owner は保持されたレシピになること。
func TestRecipeUpdate_SharedUserUpdatesRecipe(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(
		factory.WithRecipeID("r1"),
		factory.WithOwnerID("u100"),
		factory.WithSharedUsers(*factory.NewUser(factory.WithID("u7"))),
	)
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	_, err := svc.Update(context.Background(), "u7", "r1", request.RecipeRequest{Title: "更新"}) // 共有先 u7 は許可

	// Assert
	require.NoError(t, err)
	require.NotNil(t, rr.updated)
	want := domain.Recipe{
		ID:          "r1",
		Title:       "更新",
		Servings:    1,
		OwnerID:     "u100", // owner は変更されない
		Labels:      []domain.RecipeLabel{},
		SharedUsers: []domain.User{},
		Ingredients: []domain.RecipeIngredient{},
		Seasonings:  []domain.RecipeSeasoning{},
	}
	assert.Equal(t, want, *rr.updated)
}

// 自分のレシピを削除する時、対象 ID が削除されること。
func TestRecipeDelete_Success(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u5"))
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	err := svc.Delete(context.Background(), "u5", "r1")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, rr.deletedIDs, "r1")
}

// 存在しないレシピを削除する時、ErrNotFound が返ること。
func TestRecipeDelete_NotFound(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	err := svc.Delete(context.Background(), "u1", "no-such-recipe")

	// Assert
	assert.ErrorIs(t, err, ErrNotFound)
}

// 閲覧可能なレシピだけを並べ替えた時、その並びがリポジトリへ渡ること。
func TestRecipeReorder_PassesOrderToRepo(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u5"))
	rr.store["r2"] = factory.NewRecipe(factory.WithRecipeID("r2"), factory.WithOwnerID("u5"))
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	err := svc.Reorder(context.Background(), "u5", []string{"r2", "r1"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"r2", "r1"}, rr.reorderedIDs)
}

// 同じレシピ ID が重複して渡された時、重複を除いてリポジトリへ渡ること。
func TestRecipeReorder_DedupesRecipeIDs(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u5"))
	rr.store["r2"] = factory.NewRecipe(factory.WithRecipeID("r2"), factory.WithOwnerID("u5"))
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act: r1 が重複
	err := svc.Reorder(context.Background(), "u5", []string{"r1", "r2", "r1"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"r1", "r2"}, rr.reorderedIDs)
}

// 閲覧できないレシピを並べ替えに含めた時、ErrForbidden が返り保存されないこと。
func TestRecipeReorder_ForbiddenForInvisibleRecipe(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u5"))
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act: r1 は見えるが r-other は見えない
	err := svc.Reorder(context.Background(), "u5", []string{"r1", "r-other"})

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 所有レシピをアーカイブした時、その状態がリポジトリへ保存されること。
func TestRecipeSetArchived_OwnerSaves(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u5"))
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	err := svc.SetArchived(context.Background(), "u5", "r1", true)

	// Assert
	require.NoError(t, err)
	assert.True(t, rr.archived["u5"]["r1"])
}

// 共有先ユーザーは、共有されたレシピを自分の状態としてアーカイブできること。
func TestRecipeSetArchived_SharedUserAllowed(t *testing.T) {
	// Arrange: r1 は u5 所有、u9 に共有
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(
		factory.WithRecipeID("r1"),
		factory.WithOwnerID("u5"),
		factory.WithSharedUsers(*factory.NewUser(factory.WithID("u9"))),
	)
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	err := svc.SetArchived(context.Background(), "u9", "r1", true)

	// Assert
	require.NoError(t, err)
	assert.True(t, rr.archived["u9"]["r1"])
}

// 閲覧できないユーザーがアーカイブしようとした時、ErrForbidden が返り保存されないこと。
func TestRecipeSetArchived_ForbiddenForInvisibleRecipe(t *testing.T) {
	// Arrange: r1 は u5 所有・共有なし。u9 からは見えない
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u5"))
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	err := svc.SetArchived(context.Background(), "u9", "r1", true)

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 存在しないレシピをアーカイブしようとした時、ErrNotFound が返ること。
func TestRecipeSetArchived_NotFound(t *testing.T) {
	// Arrange
	rr := newMockRecipeRepo()
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	err := svc.SetArchived(context.Background(), "u1", "no-such-recipe", true)

	// Assert
	assert.ErrorIs(t, err, ErrNotFound)
}

// 更新した時、レスポンスの Archived に操作ユーザーのアーカイブ状態が反映されること。
func TestRecipeUpdate_ReflectsPerUserArchived(t *testing.T) {
	// Arrange: u5 所有の r1 を、u5 が事前にアーカイブしている
	rr := newMockRecipeRepo()
	rr.store["r1"] = factory.NewRecipe(factory.WithRecipeID("r1"), factory.WithOwnerID("u5"))
	require.NoError(t, rr.SetArchived(context.Background(), "u5", "r1", true))
	svc := NewRecipeService(rr, &mockUserRepo{})

	// Act
	updated, err := svc.Update(context.Background(), "u5", "r1", request.RecipeRequest{Title: "肉じゃが"})

	// Assert
	require.NoError(t, err)
	assert.True(t, updated.Archived)
}
