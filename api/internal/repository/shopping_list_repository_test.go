package repository

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// arrangeShoppingList は owner が所有する買い物リストを1件作成し、その ID を返す。
func arrangeShoppingList(t *testing.T) (context.Context, domain.ShoppingListRepository, *domain.User, string) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewShoppingListRepository(testDB)
	owner := seedUser(t, "owner")
	list := &domain.ShoppingList{OwnerID: owner.ID}
	require.NoError(t, repo.Create(ctx, list))
	return ctx, repo, owner, list.ID
}

// 所有する買い物リストがある時、FindForUser でそれが返ること。
func TestShoppingListRepo_FindForUser_ReturnsOwned(t *testing.T) {
	// Arrange
	ctx, repo, owner, id := arrangeShoppingList(t)

	// Act
	got, err := repo.FindForUser(ctx, owner.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
}

// 共有されたリストと自分の所有リストが両方ある時、共有された方が優先して返ること。
func TestShoppingListRepo_FindForUser_PrefersShared(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewShoppingListRepository(testDB)
	me := seedUser(t, "me")
	other := seedUser(t, "other")
	// 自分が所有する空リスト。
	own := &domain.ShoppingList{OwnerID: me.ID}
	require.NoError(t, repo.Create(ctx, own))
	// other が所有し、自分に共有したリスト。
	shared := &domain.ShoppingList{OwnerID: other.ID}
	require.NoError(t, repo.Create(ctx, shared))
	shared.SharedUsers = []domain.User{*me}
	require.NoError(t, repo.ReplaceSharedUsers(ctx, shared))

	// Act
	got, err := repo.FindForUser(ctx, me.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, shared.ID, got.ID)
}

// 見えるリストが無い時、FindForUser が nil を返すこと。
func TestShoppingListRepo_FindForUser_NoneReturnsNil(t *testing.T) {
	// Arrange
	testutil.RequireIntegration(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewShoppingListRepository(testDB)
	stranger := seedUser(t, "stranger")

	// Act
	got, err := repo.FindForUser(ctx, stranger.ID)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, got)
}

// 項目を追加した時、FindByID でチェック済みが末尾・同グループは追加順に並ぶこと。
func TestShoppingListRepo_FindByID_OrdersCheckedLast(t *testing.T) {
	// Arrange
	ctx, repo, _, id := arrangeShoppingList(t)
	first := &domain.ShoppingListItem{ShoppingListID: id, Name: "牛乳"}
	second := &domain.ShoppingListItem{ShoppingListID: id, Name: "卵"}
	third := &domain.ShoppingListItem{ShoppingListID: id, Name: "パン"}
	require.NoError(t, repo.AddItem(ctx, first))
	require.NoError(t, repo.AddItem(ctx, second))
	require.NoError(t, repo.AddItem(ctx, third))
	// 先頭の「牛乳」をチェック済みにする(末尾へ沈むはず)。
	require.NoError(t, repo.SetItemChecked(ctx, first.ID, true))

	// Act
	got, err := repo.FindByID(ctx, id)

	// Assert
	require.NoError(t, err)
	require.Len(t, got.Items, 3)
	names := []string{got.Items[0].Name, got.Items[1].Name, got.Items[2].Name}
	assert.Equal(t, []string{"卵", "パン", "牛乳"}, names)
}

// 並び替えた時、FindByID が position 昇順(チェック済みは末尾)で返すこと。
func TestShoppingListRepo_Reorder_AppliesOrder(t *testing.T) {
	// Arrange
	ctx, repo, _, id := arrangeShoppingList(t)
	a := &domain.ShoppingListItem{ShoppingListID: id, Name: "牛乳"}
	b := &domain.ShoppingListItem{ShoppingListID: id, Name: "卵"}
	c := &domain.ShoppingListItem{ShoppingListID: id, Name: "パン"}
	require.NoError(t, repo.AddItem(ctx, a))
	require.NoError(t, repo.AddItem(ctx, b))
	require.NoError(t, repo.AddItem(ctx, c))

	// Act: パン → 牛乳 → 卵 の順に並べ替える
	require.NoError(t, repo.Reorder(ctx, id, []string{c.ID, a.ID, b.ID}))

	// Assert
	got, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Len(t, got.Items, 3)
	names := []string{got.Items[0].Name, got.Items[1].Name, got.Items[2].Name}
	assert.Equal(t, []string{"パン", "牛乳", "卵"}, names)
}

// 追加した項目が末尾の position を採番され、追加順で並ぶこと。
func TestShoppingListRepo_AddItem_AppendsToEnd(t *testing.T) {
	// Arrange
	ctx, repo, _, id := arrangeShoppingList(t)
	first := &domain.ShoppingListItem{ShoppingListID: id, Name: "牛乳"}
	require.NoError(t, repo.AddItem(ctx, first))
	require.NoError(t, repo.Reorder(ctx, id, []string{first.ID})) // position を 0 に確定

	// Act: 追加後の項目は末尾(position > 既存)へ回る
	second := &domain.ShoppingListItem{ShoppingListID: id, Name: "卵"}
	require.NoError(t, repo.AddItem(ctx, second))

	// Assert
	got, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Len(t, got.Items, 2)
	assert.Equal(t, []string{"牛乳", "卵"}, []string{got.Items[0].Name, got.Items[1].Name})
}

// チェック済みを一括削除した時、未チェックの項目だけが残ること。
func TestShoppingListRepo_DeleteCheckedItems_KeepsUnchecked(t *testing.T) {
	// Arrange
	ctx, repo, _, id := arrangeShoppingList(t)
	checked := &domain.ShoppingListItem{ShoppingListID: id, Name: "牛乳"}
	unchecked := &domain.ShoppingListItem{ShoppingListID: id, Name: "卵"}
	require.NoError(t, repo.AddItem(ctx, checked))
	require.NoError(t, repo.AddItem(ctx, unchecked))
	require.NoError(t, repo.SetItemChecked(ctx, checked.ID, true))

	// Act
	require.NoError(t, repo.DeleteCheckedItems(ctx, id))

	// Assert
	got, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "卵", got.Items[0].Name)
}
