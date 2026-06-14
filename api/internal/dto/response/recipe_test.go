package response

import (
	"testing"
	"time"

	"recipe-backend/internal/domain"
)

func TestToRecipeResponse_MapsAndFormats(t *testing.T) {
	createTime := 30
	// 2026-06-14 03:00 UTC → JST(+9) では 12:00
	created := time.Date(2026, 6, 14, 3, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 6, 14, 4, 30, 0, 0, time.UTC) // JST 13:30

	r := &domain.Recipe{
		ID:         1,
		Title:      "肉じゃが",
		CreateTime: &createTime,
		CreateFor:  2,
		Procedure:  "煮る",
		ArchiveFlg: false,
		OwnerID:    10,
		Owner:      domain.ApplicationUser{ID: 10, Username: "alice"},
		Labels:     []domain.RecipeLabel{{ID: 1, Name: "和食"}},
		SharedUsers: []domain.ApplicationUser{
			{ID: 20, Username: "bob"},
		},
		Cooking: []domain.Cooking{
			{Ingredient: domain.Ingredient{ID: 5, Name: "じゃがいも"}, Quantity: 3, Unit: "個"},
		},
		Season: []domain.Season{
			{Seasoning: domain.Seasoning{ID: 7, Name: "醤油"}, Quantity: 2, Unit: "大さじ"},
		},
		CreatedAt: created,
		UpdatedAt: updated,
	}

	got := ToRecipeResponse(r)

	if got.CreatedAt != "2026-06-14 12:00" {
		t.Errorf("created_at = %q, want %q (JST)", got.CreatedAt, "2026-06-14 12:00")
	}
	if got.UpdatedAt != "2026-06-14 13:30" {
		t.Errorf("updated_at = %q, want %q (JST)", got.UpdatedAt, "2026-06-14 13:30")
	}
	if got.Owner.ID != 10 || got.Owner.Username != "alice" {
		t.Errorf("owner = %+v, want {10 alice}", got.Owner)
	}
	if len(got.Label) != 1 || got.Label[0].Name != "和食" {
		t.Errorf("label = %+v", got.Label)
	}
	if len(got.SharedUser) != 1 || got.SharedUser[0].Username != "bob" {
		t.Errorf("shared_user = %+v", got.SharedUser)
	}
	if len(got.Cooking) != 1 || got.Cooking[0].Ingredients.Name != "じゃがいも" || got.Cooking[0].Quantity != 3 {
		t.Errorf("cooking = %+v", got.Cooking)
	}
	if len(got.Season) != 1 || got.Season[0].Seasoning.Name != "醤油" || got.Season[0].Unit != "大さじ" {
		t.Errorf("season = %+v", got.Season)
	}
	if got.CreateTime == nil || *got.CreateTime != 30 {
		t.Errorf("create_time = %v, want 30", got.CreateTime)
	}
}

func TestToRecipeResponse_EmptySlicesNotNil(t *testing.T) {
	// 関連が空でも JSON では [] になるよう、nil ではなく空スライスを返すこと。
	r := &domain.Recipe{ID: 1, Owner: domain.ApplicationUser{ID: 1}}
	got := ToRecipeResponse(r)

	if got.Cooking == nil || got.Season == nil || got.Label == nil || got.SharedUser == nil {
		t.Errorf("empty associations must be non-nil slices: cooking=%v season=%v label=%v shared=%v",
			got.Cooking, got.Season, got.Label, got.SharedUser)
	}
}
