package service

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 自分の所有リストが既にある時、Get でそれが返ること。
func TestShoppingListGet_ReturnsOwned(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1"}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	list, err := svc.Get(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "l1", list.ID)
}

// 見えるリストがまだ無い時、Get で自分が所有する空のリストが作られること。
func TestShoppingListGet_CreatesWhenNone(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.Get(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, lr.created)
	assert.Equal(t, "u1", lr.created.OwnerID)
}

// 自分に共有されたリストと自分の所有リストが両方ある時、共有された方が優先して返ること。
func TestShoppingListGet_PrefersShared(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["own"] = &domain.ShoppingList{ID: "own", OwnerID: "u1"}
	lr.store["shared"] = &domain.ShoppingList{ID: "shared", OwnerID: "u2", SharedUsers: []domain.User{{ID: "u1"}}}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	list, err := svc.Get(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "shared", list.ID)
}

// 所有者が項目を追加した時、名前付きでリポジトリに保存されること。
func TestShoppingListAddItem_Saves(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1"}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.AddItem(context.Background(), "u1", "l1", "牛乳")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, lr.addedItem)
	assert.Equal(t, "牛乳", lr.addedItem.Name)
}

// 共有相手が項目を追加した時、許可されること(owner でなくても共有先なら操作できる)。
func TestShoppingListAddItem_AllowedForSharedUser(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u2", SharedUsers: []domain.User{{ID: "u1"}}}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.AddItem(context.Background(), "u1", "l1", "卵")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "卵", lr.addedItem.Name)
}

// 所有でも共有先でもないユーザーが項目を追加しようとした時、ErrForbidden が返ること。
func TestShoppingListAddItem_Forbidden(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u2"}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.AddItem(context.Background(), "u1", "l1", "牛乳")

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 存在しないリストに項目を追加しようとした時、ErrNotFound が返ること。
func TestShoppingListAddItem_NotFound(t *testing.T) {
	// Arrange
	svc := NewShoppingListService(newMockShoppingListRepo(), &mockUserRepo{})

	// Act
	_, err := svc.AddItem(context.Background(), "u1", "no-such", "牛乳")

	// Assert
	assert.ErrorIs(t, err, ErrNotFound)
}

// 項目をチェックした時、チェック状態がリポジトリに反映されること。
func TestShoppingListSetItemChecked_Updates(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1", Items: []domain.ShoppingListItem{{ID: "i1", Name: "牛乳"}}}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.SetItemChecked(context.Background(), "u1", "l1", "i1", true)

	// Assert
	require.NoError(t, err)
	assert.True(t, lr.checkedItems["i1"])
}

// リストに属さない項目 ID を指定した時、ErrNotFound が返ること。
func TestShoppingListSetItemChecked_ItemNotInList(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1", Items: []domain.ShoppingListItem{{ID: "i1", Name: "牛乳"}}}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.SetItemChecked(context.Background(), "u1", "l1", "other", true)

	// Assert
	assert.ErrorIs(t, err, ErrNotFound)
}

// 項目を削除した時、リポジトリから削除されること。
func TestShoppingListDeleteItem_Removes(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1", Items: []domain.ShoppingListItem{{ID: "i1", Name: "牛乳"}}}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.DeleteItem(context.Background(), "u1", "l1", "i1")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, lr.deletedItems, "i1")
}

// チェック済みを一括削除した時、リポジトリの一括削除が呼ばれること。
func TestShoppingListClearChecked_Clears(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1"}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.ClearChecked(context.Background(), "u1", "l1")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, lr.clearedLists, "l1")
}

// 項目を並び替えた時、指定した順序がリポジトリに渡ること。
func TestShoppingListReorder_PassesOrderToRepo(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1", Items: []domain.ShoppingListItem{
		{ID: "i1", Name: "牛乳"}, {ID: "i2", Name: "卵"},
	}}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.Reorder(context.Background(), "u1", "l1", []string{"i2", "i1"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"i2", "i1"}, lr.reorderedIDs)
}

// このリストに属さない項目 ID を並び替えに含めた時、ErrNotFound が返ること。
func TestShoppingListReorder_ItemNotInList(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1", Items: []domain.ShoppingListItem{{ID: "i1", Name: "牛乳"}}}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.Reorder(context.Background(), "u1", "l1", []string{"i1", "other"})

	// Assert
	assert.ErrorIs(t, err, ErrNotFound)
}

// 所有でも共有先でもないユーザーが並び替えようとした時、ErrForbidden が返ること。
func TestShoppingListReorder_Forbidden(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u2", Items: []domain.ShoppingListItem{{ID: "i1", Name: "牛乳"}}}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.Reorder(context.Background(), "u1", "l1", []string{"i1"})

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 共有相手を更新した時、username から解決したユーザーがリストに設定されること。
func TestShoppingListUpdateShares_ResolvesUsers(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1"}
	ur := &mockUserRepo{byName: map[string]*domain.User{"partner": {ID: "u2", Username: "partner"}}}
	svc := NewShoppingListService(lr, ur)

	// Act
	_, err := svc.UpdateShares(context.Background(), "u1", "l1", []request.SharedUserInput{{Username: "partner"}})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, lr.sharesReplace)
	require.Len(t, lr.sharesReplace.SharedUsers, 1)
	assert.Equal(t, "u2", lr.sharesReplace.SharedUsers[0].ID)
}

// 存在しない username を共有相手に指定した時、ErrSharedUserNotFound が返ること。
func TestShoppingListUpdateShares_SharedUserNotFound(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1"}
	svc := NewShoppingListService(lr, &mockUserRepo{})

	// Act
	_, err := svc.UpdateShares(context.Background(), "u1", "l1", []request.SharedUserInput{{Username: "ghost"}})

	// Assert
	assert.ErrorIs(t, err, ErrSharedUserNotFound)
}
