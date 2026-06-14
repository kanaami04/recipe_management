package repository

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// arrangeLabelRepo は結合テストの前提を整え、空の LabelRepository を返す。
func arrangeLabelRepo(t *testing.T) (context.Context, domain.LabelRepository) {
	t.Helper()
	testutil.RequireIntegration(t)
	truncateAll(t)
	return context.Background(), NewLabelRepository(testDB)
}

// ラベルが複数登録されている時、FindAll で ID 昇順のまま構造体ごと全件返ること。
func TestLabelRepo_FindAll_ReturnsAllOrderedByID(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)
	require.NoError(t, testDB.Create(&domain.RecipeLabel{Name: "和食"}).Error)
	require.NoError(t, testDB.Create(&domain.RecipeLabel{Name: "洋食"}).Error)
	require.NoError(t, testDB.Create(&domain.RecipeLabel{Name: "中華"}).Error)

	// Act
	got, err := repo.FindAll(ctx)

	// Assert: ID 昇順で全件返ること
	require.NoError(t, err)
	assert.Equal(t, []domain.RecipeLabel{
		{ID: 1, Name: "和食"},
		{ID: 2, Name: "洋食"},
		{ID: 3, Name: "中華"},
	}, got)
}

// ラベルが1件もない時、FindAll で空が返ること。
func TestLabelRepo_FindAll_Empty(t *testing.T) {
	// Arrange
	ctx, repo := arrangeLabelRepo(t)

	// Act
	got, err := repo.FindAll(ctx)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, got)
}
