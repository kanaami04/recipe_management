package repository

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil"
)

func TestRecipeRepo_CreateAndFindByID(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)

	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	label, _ := repo.GetOrCreateLabel(ctx, "和食")
	ing, _ := repo.GetOrCreateIngredient(ctx, "じゃがいも")
	sea, _ := repo.GetOrCreateSeasoning(ctx, "醤油")

	r := &domain.Recipe{
		Title:       "肉じゃが",
		CreateFor:   2,
		OwnerID:     owner.ID,
		Labels:      []domain.RecipeLabel{*label},
		SharedUsers: []domain.ApplicationUser{*friend},
		Cooking:     []domain.Cooking{{IngredientID: ing.ID, Quantity: 3, Unit: "個"}},
		Season:      []domain.Season{{SeasoningID: sea.ID, Quantity: 2, Unit: "大さじ"}},
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.FindByID(ctx, r.ID)
	if err != nil || got == nil {
		t.Fatalf("findByID: got=%v err=%v", got, err)
	}
	if got.Owner.Username != "owner" {
		t.Errorf("owner = %q, want owner", got.Owner.Username)
	}
	if len(got.Labels) != 1 || got.Labels[0].Name != "和食" {
		t.Errorf("labels = %+v", got.Labels)
	}
	if len(got.SharedUsers) != 1 || got.SharedUsers[0].Username != "friend" {
		t.Errorf("shared = %+v", got.SharedUsers)
	}
	if len(got.Cooking) != 1 || got.Cooking[0].Ingredient.Name != "じゃがいも" || got.Cooking[0].Quantity != 3 {
		t.Errorf("cooking = %+v", got.Cooking)
	}
	if len(got.Season) != 1 || got.Season[0].Seasoning.Name != "醤油" {
		t.Errorf("season = %+v", got.Season)
	}
}

func TestRecipeRepo_GetOrCreate_ReusesByName(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)

	a, _ := repo.GetOrCreateLabel(ctx, "和食")
	b, _ := repo.GetOrCreateLabel(ctx, "和食")
	if a.ID != b.ID {
		t.Errorf("same label name should reuse row: a=%d b=%d", a.ID, b.ID)
	}

	i1, _ := repo.GetOrCreateIngredient(ctx, "塩")
	i2, _ := repo.GetOrCreateIngredient(ctx, "塩")
	if i1.ID != i2.ID {
		t.Errorf("same ingredient name should reuse row: %d vs %d", i1.ID, i2.ID)
	}
}

func TestRecipeRepo_FindAllForUser_Filtering(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)

	owner := seedUser(t, "owner")
	friend := seedUser(t, "friend")
	stranger := seedUser(t, "stranger")

	r := &domain.Recipe{
		Title:       "共有レシピ",
		CreateFor:   1,
		OwnerID:     owner.ID,
		SharedUsers: []domain.ApplicationUser{*friend},
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("create: %v", err)
	}

	ownerList, _ := repo.FindAllForUser(ctx, owner.ID)
	if len(ownerList) != 1 {
		t.Errorf("owner should see 1 recipe, got %d", len(ownerList))
	}
	friendList, _ := repo.FindAllForUser(ctx, friend.ID)
	if len(friendList) != 1 {
		t.Errorf("shared user should see 1 recipe, got %d", len(friendList))
	}
	strangerList, _ := repo.FindAllForUser(ctx, stranger.ID)
	if len(strangerList) != 0 {
		t.Errorf("stranger should see 0 recipes, got %d", len(strangerList))
	}
}

func TestRecipeRepo_Update_ReplaceSemantics(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)

	owner := seedUser(t, "owner")
	ingA, _ := repo.GetOrCreateIngredient(ctx, "じゃがいも")
	labelA, _ := repo.GetOrCreateLabel(ctx, "和食")

	r := &domain.Recipe{
		Title:     "初版",
		CreateFor: 1,
		OwnerID:   owner.ID,
		Labels:    []domain.RecipeLabel{*labelA},
		Cooking:   []domain.Cooking{{IngredientID: ingA.ID, Quantity: 1, Unit: "個"}},
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("create: %v", err)
	}

	// 別の食材・別ラベルに差し替え
	ingB, _ := repo.GetOrCreateIngredient(ctx, "人参")
	labelB, _ := repo.GetOrCreateLabel(ctx, "夕食")
	r.Title = "改訂版"
	r.Labels = []domain.RecipeLabel{*labelB}
	r.Cooking = []domain.Cooking{{IngredientID: ingB.ID, Quantity: 2, Unit: "本"}}
	r.Season = nil
	if err := repo.Update(ctx, r); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := repo.FindByID(ctx, r.ID)
	if got.Title != "改訂版" {
		t.Errorf("title = %q, want 改訂版", got.Title)
	}
	if len(got.Cooking) != 1 || got.Cooking[0].Ingredient.Name != "人参" {
		t.Errorf("cooking should be replaced to 人参: %+v", got.Cooking)
	}
	if len(got.Labels) != 1 || got.Labels[0].Name != "夕食" {
		t.Errorf("labels should be replaced to 夕食: %+v", got.Labels)
	}
}

func TestRecipeRepo_Delete(t *testing.T) {
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewRecipeRepository(testDB)

	owner := seedUser(t, "owner")
	ing, _ := repo.GetOrCreateIngredient(ctx, "卵")
	r := &domain.Recipe{
		Title:     "削除対象",
		CreateFor: 1,
		OwnerID:   owner.ID,
		Cooking:   []domain.Cooking{{IngredientID: ing.ID, Quantity: 2, Unit: "個"}},
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.Delete(ctx, r); err != nil {
		t.Fatalf("delete: %v", err)
	}

	got, _ := repo.FindByID(ctx, r.ID)
	if got != nil {
		t.Errorf("recipe should be deleted, got %+v", got)
	}

	// 子テーブル(cooking)も消えていること
	var cookingCount int64
	testDB.Model(&domain.Cooking{}).Where("recipe_id = ?", r.ID).Count(&cookingCount)
	if cookingCount != 0 {
		t.Errorf("cooking rows should be deleted, got %d", cookingCount)
	}
}
