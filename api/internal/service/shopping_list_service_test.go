package service

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// グループ未所属で自分の所有リストがある時、Get でそれが返ること。
func TestShoppingListGet_ReturnsOwned(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1"}
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

	// Act
	list, err := svc.Get(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "l1", list.ID)
}

// リストがまだ無い時、Get で自分が所有する空のリストが作られること。
func TestShoppingListGet_CreatesWhenNone(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

	// Act
	_, err := svc.Get(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, lr.created)
	assert.Equal(t, "u1", lr.created.OwnerID)
}

// グループ所属時、Get でグループ所有者のリスト(= グループの 1 リスト)が返ること。
func TestShoppingListGet_UsesGroupOwnersList(t *testing.T) {
	// Arrange: グループ所有者 uOwner のリスト。uMember が取得する。
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "uOwner"}
	gr := newMockShareGroupRepo()
	gr.seed("g1", "uOwner", "uMember")
	svc := NewShoppingListService(lr, gr)

	// Act
	list, err := svc.Get(context.Background(), "uMember")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "l1", list.ID)
}

// 所有者が項目を追加した時、名前付きでリポジトリに保存されること。
func TestShoppingListAddItem_Saves(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u1"}
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

	// Act
	_, err := svc.AddItem(context.Background(), "u1", "l1", "牛乳")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, lr.addedItem)
	assert.Equal(t, "牛乳", lr.addedItem.Name)
}

// 同じグループのメンバーが項目を追加した時、許可されること。
func TestShoppingListAddItem_AllowedForGroupMember(t *testing.T) {
	// Arrange: l1 は uOwner 所有。uMember は同じグループ。
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "uOwner"}
	gr := newMockShareGroupRepo()
	gr.seed("g1", "uOwner", "uMember")
	svc := NewShoppingListService(lr, gr)

	// Act
	_, err := svc.AddItem(context.Background(), "uMember", "l1", "卵")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "卵", lr.addedItem.Name)
}

// グループ外のユーザーが項目を追加しようとした時、ErrForbidden が返ること。
func TestShoppingListAddItem_Forbidden(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u2"}
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

	// Act
	_, err := svc.AddItem(context.Background(), "u1", "l1", "牛乳")

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 存在しないリストに項目を追加しようとした時、ErrNotFound が返ること。
func TestShoppingListAddItem_NotFound(t *testing.T) {
	// Arrange
	svc := NewShoppingListService(newMockShoppingListRepo(), newMockShareGroupRepo())

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
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

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
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

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
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

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
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

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
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

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
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

	// Act
	_, err := svc.Reorder(context.Background(), "u1", "l1", []string{"i1", "other"})

	// Assert
	assert.ErrorIs(t, err, ErrNotFound)
}

// グループ外のユーザーが並び替えようとした時、ErrForbidden が返ること。
func TestShoppingListReorder_Forbidden(t *testing.T) {
	// Arrange
	lr := newMockShoppingListRepo()
	lr.store["l1"] = &domain.ShoppingList{ID: "l1", OwnerID: "u2", Items: []domain.ShoppingListItem{{ID: "i1", Name: "牛乳"}}}
	svc := NewShoppingListService(lr, newMockShareGroupRepo())

	// Act
	_, err := svc.Reorder(context.Background(), "u1", "l1", []string{"i1"})

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}
