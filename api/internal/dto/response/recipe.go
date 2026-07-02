package response

import (
	"time"

	"recipe-backend/internal/apigen"
	"recipe-backend/internal/domain"
)

// 構造体定義は openapi.yaml から生成する。生成型を再エクスポートする。
type (
	NameResponse    = apigen.NameResponse
	LabelResponse   = apigen.LabelResponse
	CookingResponse = apigen.CookingResponse
	SeasonResponse  = apigen.SeasonResponse
	RecipeResponse  = apigen.RecipeResponse
)

// jst は created_at/updated_at の整形に使うタイムゾーン。
var jst = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.FixedZone("JST", 9*60*60)
	}
	return loc
}()

const dateLayout = "2006-01-02 15:04" // DRF の "%Y-%m-%d %H:%M" 相当

// ToRecipeResponse は domain.Recipe を API 契約に合わせた DTO へ変換する。
func ToRecipeResponse(r *domain.Recipe) RecipeResponse {
	cooking := make([]CookingResponse, 0, len(r.Ingredients))
	for i := range r.Ingredients {
		ing := &r.Ingredients[i]
		cooking = append(cooking, CookingResponse{
			Ingredients: NameResponse{Name: ing.Name},
			Quantity:    ing.Quantity,
			Unit:        ing.Unit,
		})
	}

	season := make([]SeasonResponse, 0, len(r.Seasonings))
	for i := range r.Seasonings {
		sea := &r.Seasonings[i]
		season = append(season, SeasonResponse{
			Seasoning: NameResponse{Name: sea.Name},
			Quantity:  sea.Quantity,
			Unit:      sea.Unit,
		})
	}

	labels := make([]LabelResponse, 0, len(r.Labels))
	for i := range r.Labels {
		labels = append(labels, LabelResponse{Name: r.Labels[i].Name})
	}

	shared := make([]UserListItem, 0, len(r.SharedUsers))
	for i := range r.SharedUsers {
		shared = append(shared, UserListItem{ID: r.SharedUsers[i].ID, Username: r.SharedUsers[i].Username})
	}

	return RecipeResponse{
		ID:         r.ID,
		CreatedAt:  r.CreatedAt.In(jst).Format(dateLayout),
		UpdatedAt:  r.UpdatedAt.In(jst).Format(dateLayout),
		Cooking:    cooking,
		Season:     season,
		Procedure:  r.Procedure,
		Owner:      UserListItem{ID: r.Owner.ID, Username: r.Owner.Username},
		SharedUser: shared,
		Title:      r.Title,
		CreateTime: r.CookingTime,
		CreateFor:  r.Servings,
		ArchiveFlg: r.Archived,
		Label:      labels,
	}
}
