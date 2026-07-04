package service

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ラベルを作成した時、owner 付きでリポジトリに保存されること。
func TestLabelCreate_Saves(t *testing.T) {
	// Arrange
	lr := newMockLabelRepo()
	svc := NewLabelService(lr)

	// Act
	_, err := svc.Create(context.Background(), "u1", "和食")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, lr.created)
	assert.Equal(t, domain.Label{ID: lr.created.ID, Name: "和食", OwnerID: "u1"}, *lr.created)
}

// 同名ラベルが既にある時、作成すると ErrDuplicate が返ること。
func TestLabelCreate_Duplicate(t *testing.T) {
	// Arrange
	lr := newMockLabelRepo()
	lr.store["l1"] = &domain.Label{ID: "l1", Name: "和食", OwnerID: "u1"}
	svc := NewLabelService(lr)

	// Act
	_, err := svc.Create(context.Background(), "u1", "和食")

	// Assert
	assert.ErrorIs(t, err, ErrDuplicate)
}

// 自分のラベルを改名した時、新しい名前のラベルが返ること。
func TestLabelRename_ReturnsRenamed(t *testing.T) {
	// Arrange
	lr := newMockLabelRepo()
	lr.store["l1"] = &domain.Label{ID: "l1", Name: "和食", OwnerID: "u1"}
	svc := NewLabelService(lr)

	// Act
	label, err := svc.Rename(context.Background(), "u1", "l1", "日本料理")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "日本料理", label.Name)
}

// 自分のラベルを改名した時、新しい名前がリポジトリに渡ること。
func TestLabelRename_PassesNewNameToRepo(t *testing.T) {
	// Arrange
	lr := newMockLabelRepo()
	lr.store["l1"] = &domain.Label{ID: "l1", Name: "和食", OwnerID: "u1"}
	svc := NewLabelService(lr)

	// Act
	_, err := svc.Rename(context.Background(), "u1", "l1", "日本料理")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "日本料理", lr.renamedTo)
}

// 改名先の名前が既にある時、ErrDuplicate が返ること。
func TestLabelRename_Duplicate(t *testing.T) {
	// Arrange
	lr := newMockLabelRepo()
	lr.store["l1"] = &domain.Label{ID: "l1", Name: "和食", OwnerID: "u1"}
	lr.store["l2"] = &domain.Label{ID: "l2", Name: "洋食", OwnerID: "u1"}
	svc := NewLabelService(lr)

	// Act: l1 を既存の「洋食」へ改名
	_, err := svc.Rename(context.Background(), "u1", "l1", "洋食")

	// Assert
	assert.ErrorIs(t, err, ErrDuplicate)
}

// 他人のラベルを改名しようとした時、ErrForbidden が返ること。
func TestLabelRename_Forbidden(t *testing.T) {
	// Arrange: l1 は u2 所有
	lr := newMockLabelRepo()
	lr.store["l1"] = &domain.Label{ID: "l1", Name: "和食", OwnerID: "u2"}
	svc := NewLabelService(lr)

	// Act
	_, err := svc.Rename(context.Background(), "u1", "l1", "日本料理")

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 存在しないラベルを改名しようとした時、ErrNotFound が返ること。
func TestLabelRename_NotFound(t *testing.T) {
	// Arrange
	svc := NewLabelService(newMockLabelRepo())

	// Act
	_, err := svc.Rename(context.Background(), "u1", "no-such", "日本料理")

	// Assert
	assert.ErrorIs(t, err, ErrNotFound)
}

// 自分のラベルを削除した時、リポジトリから削除されること。
func TestLabelDelete_Removes(t *testing.T) {
	// Arrange
	lr := newMockLabelRepo()
	lr.store["l1"] = &domain.Label{ID: "l1", Name: "和食", OwnerID: "u1"}
	svc := NewLabelService(lr)

	// Act
	err := svc.Delete(context.Background(), "u1", "l1")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, lr.deletedIDs, "l1")
}

// 他人のラベルを削除しようとした時、ErrForbidden が返ること。
func TestLabelDelete_Forbidden(t *testing.T) {
	// Arrange: l1 は u2 所有
	lr := newMockLabelRepo()
	lr.store["l1"] = &domain.Label{ID: "l1", Name: "和食", OwnerID: "u2"}
	svc := NewLabelService(lr)

	// Act
	err := svc.Delete(context.Background(), "u1", "l1")

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// ラベルが登録されている時、List で自分のラベルが返ること。
func TestLabelList_ReturnsOwn(t *testing.T) {
	// Arrange
	lr := newMockLabelRepo()
	lr.store["l1"] = &domain.Label{ID: "l1", Name: "和食", OwnerID: "u1"}
	svc := NewLabelService(lr)

	// Act
	labels, err := svc.List(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	require.Len(t, labels, 1)
	assert.Equal(t, "和食", labels[0].Name)
}
