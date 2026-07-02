package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ラベルが登録されている時、List で取得順のまま全件が返ること。
func TestLabelList_ReturnsAll(t *testing.T) {
	// Arrange
	lr := &mockLabelRepo{names: []string{"和食", "洋食"}}
	svc := NewLabelService(lr)

	// Act
	names, err := svc.List(context.Background(), 1)

	// Assert: 取得順に全件返ること
	require.NoError(t, err)
	assert.Equal(t, []string{"和食", "洋食"}, names)
}

// ラベルが1件も無い時、List で空が返ること。
func TestLabelList_Empty(t *testing.T) {
	// Arrange
	svc := NewLabelService(&mockLabelRepo{})

	// Act
	names, err := svc.List(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, names)
}
