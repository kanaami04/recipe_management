package response

import (
	"testing"
	"time"

	"recipe-backend/internal/domain"

	"github.com/stretchr/testify/assert"
)

// レシピを変換した時、各項目がマッピングされ日時が JST 文字列に整形された DTO になること。
func TestToRecipeResponse_MapsAndFormats(t *testing.T) {
	// Arrange
	createTime := 30
	// 2026-06-14 03:00 UTC → JST(+9) では 12:00
	created := time.Date(2026, 6, 14, 3, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 6, 14, 4, 30, 0, 0, time.UTC) // JST 13:30
	r := &domain.Recipe{
		ID:          1,
		Title:       "肉じゃが",
		CreateTime:  &createTime,
		CreateFor:   2,
		Procedure:   "煮る",
		ArchiveFlg:  false,
		OwnerID:     10,
		Owner:       domain.ApplicationUser{ID: 10, Username: "alice"},
		Labels:      []domain.RecipeLabel{{ID: 1, Name: "和食"}},
		SharedUsers: []domain.ApplicationUser{{ID: 20, Username: "bob"}},
		Cooking: []domain.Cooking{
			{Ingredient: domain.Ingredient{ID: 5, Name: "じゃがいも"}, Quantity: 3, Unit: "個"},
		},
		Season: []domain.Season{
			{Seasoning: domain.Seasoning{ID: 7, Name: "醤油"}, Quantity: 2, Unit: "大さじ"},
		},
		CreatedAt: created,
		UpdatedAt: updated,
	}

	// Act
	got := ToRecipeResponse(r)

	// Assert
	want := RecipeResponse{
		ID:         1,
		CreatedAt:  "2026-06-14 12:00",
		UpdatedAt:  "2026-06-14 13:30",
		Cooking:    []CookingResponse{{Ingredients: NamedRef{ID: 5, Name: "じゃがいも"}, Quantity: 3, Unit: "個"}},
		Season:     []SeasonResponse{{Seasoning: NamedRef{ID: 7, Name: "醤油"}, Quantity: 2, Unit: "大さじ"}},
		Procedure:  "煮る",
		Owner:      UserListItem{ID: 10, Username: "alice"},
		SharedUser: []UserListItem{{ID: 20, Username: "bob"}},
		Title:      "肉じゃが",
		CreateTime: &createTime,
		CreateFor:  2,
		ArchiveFlg: false,
		Label:      []LabelResponse{{ID: 1, Name: "和食"}},
	}
	assert.Equal(t, want, got)
}

// 関連が空のレシピを変換した時、各スライスが nil ではなく空スライス([])になること。
func TestToRecipeResponse_EmptySlicesNotNil(t *testing.T) {
	// Arrange
	r := &domain.Recipe{ID: 1, Owner: domain.ApplicationUser{ID: 1}}

	// Act
	got := ToRecipeResponse(r)

	// Assert: JSON で [] になるよう、nil ではなく空スライスを返すこと。
	assert.NotNil(t, got.Cooking)
	assert.NotNil(t, got.Season)
	assert.NotNil(t, got.Label)
	assert.NotNil(t, got.SharedUser)
}
