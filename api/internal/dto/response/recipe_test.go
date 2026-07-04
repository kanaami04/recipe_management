package response

import (
	"context"
	"testing"
	"time"

	"recipe-backend/internal/domain"

	"github.com/stretchr/testify/assert"
)

// mockAvatarStorage は avatar_url 組み立て用の AvatarStorage モック。PublicURL は key を URL 化する。
type mockAvatarStorage struct{}

func (mockAvatarStorage) PresignUpload(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (mockAvatarStorage) Delete(_ context.Context, _ string) error { return nil }
func (mockAvatarStorage) PublicURL(key string) string              { return "https://cdn.example/" + key }

// レシピを変換した時、各項目がマッピングされ日時が JST 文字列に整形された DTO になること。
func TestToRecipeResponse_MapsAndFormats(t *testing.T) {
	// Arrange
	cookingTime := 30
	// 2026-06-14 03:00 UTC → JST(+9) では 12:00
	created := time.Date(2026, 6, 14, 3, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 6, 14, 4, 30, 0, 0, time.UTC) // JST 13:30
	r := &domain.Recipe{
		ID:          "r1",
		Title:       "肉じゃが",
		CookingTime: &cookingTime,
		Servings:    2,
		Procedure:   "煮る",
		Archived:    false,
		OwnerID:     "u10",
		Owner:       domain.User{ID: "u10", Username: "alice"},
		Labels:      []domain.RecipeLabel{{ID: "l1", RecipeID: "r1", Name: "和食"}},
		SharedUsers: []domain.User{{ID: "u20", Username: "bob"}},
		Ingredients: []domain.RecipeIngredient{
			{ID: "i5", RecipeID: "r1", Name: "じゃがいも", Quantity: 3, Unit: "個"},
		},
		Seasonings: []domain.RecipeSeasoning{
			{ID: "s7", RecipeID: "r1", Name: "醤油", Quantity: 0.5, Unit: "大さじ"}, // 小数(大さじ1/2)が保たれること
		},
		CreatedAt: created,
		UpdatedAt: updated,
	}

	// Act
	got := ToRecipeResponse(r, mockAvatarStorage{})

	// Assert
	want := RecipeResponse{
		ID:         "r1",
		CreatedAt:  "2026-06-14 12:00",
		UpdatedAt:  "2026-06-14 13:30",
		Cooking:    []CookingResponse{{Ingredients: NameResponse{Name: "じゃがいも"}, Quantity: 3, Unit: "個"}},
		Season:     []SeasonResponse{{Seasoning: NameResponse{Name: "醤油"}, Quantity: 0.5, Unit: "大さじ"}},
		Procedure:  "煮る",
		Owner:      UserListItem{ID: "u10", Username: "alice"},
		SharedUser: []UserListItem{{ID: "u20", Username: "bob"}},
		Title:      "肉じゃが",
		CreateTime: &cookingTime,
		CreateFor:  2,
		ArchiveFlg: false,
		Label:      []LabelResponse{{Name: "和食"}},
	}
	assert.Equal(t, want, got)
}

// 共有ユーザーにアバターがある時、shared_user の avatar_url が PublicURL で組まれること。
func TestToRecipeResponse_SharedUserAvatarURL(t *testing.T) {
	// Arrange
	key := "avatars/u20/pic.png"
	r := &domain.Recipe{
		ID:          "r1",
		Owner:       domain.User{ID: "u10", Username: "alice"},
		SharedUsers: []domain.User{{ID: "u20", Username: "bob", AvatarKey: &key}},
	}

	// Act
	got := ToRecipeResponse(r, mockAvatarStorage{})

	// Assert
	wantURL := "https://cdn.example/" + key
	assert.Equal(t, &wantURL, got.SharedUser[0].AvatarUrl)
}

// 関連が空のレシピを変換した時、各スライスが nil ではなく空スライス([])になること。
func TestToRecipeResponse_EmptySlicesNotNil(t *testing.T) {
	// Arrange
	r := &domain.Recipe{ID: "r1", Owner: domain.User{ID: "u1"}}

	// Act
	got := ToRecipeResponse(r, mockAvatarStorage{})

	// Assert: JSON で [] になるよう、nil ではなく空スライスを返すこと。
	assert.NotNil(t, got.Cooking)
	assert.NotNil(t, got.Season)
	assert.NotNil(t, got.Label)
	assert.NotNil(t, got.SharedUser)
}
