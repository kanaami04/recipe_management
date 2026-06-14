package service

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ラベルが登録されている時、List で取得順のまま全件が返ること。
func TestLabelList_ReturnsAll(t *testing.T) {
	// Arrange
	lr := &mockLabelRepo{labels: []domain.RecipeLabel{
		factory.NewRecipeLabel("和食"),
		factory.NewRecipeLabel("洋食"),
	}}
	svc := NewLabelService(lr)

	// Act
	labels, err := svc.List(context.Background())

	// Assert: 取得順に全件返ること
	require.NoError(t, err)
	names := make([]string, len(labels))
	for i, l := range labels {
		names[i] = l.Name
	}
	assert.Equal(t, []string{"和食", "洋食"}, names)
}

// ラベルが1件も無い時、List で空が返ること。
func TestLabelList_Empty(t *testing.T) {
	// Arrange
	svc := NewLabelService(&mockLabelRepo{})

	// Act
	labels, err := svc.List(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Empty(t, labels)
}
