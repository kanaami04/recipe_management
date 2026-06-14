package response

import (
	"time"

	"recipe-backend/internal/domain"
)

type NamedRef struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type LabelResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CookingResponse struct {
	Ingredients NamedRef `json:"ingredients"`
	Quantity    int      `json:"quantity"`
	Unit        string   `json:"unit"`
}

type SeasonResponse struct {
	Seasoning NamedRef `json:"seasoning"`
	Quantity  int      `json:"quantity"`
	Unit      string   `json:"unit"`
}

type RecipeResponse struct {
	ID         uint              `json:"id"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
	Cooking    []CookingResponse `json:"cooking"`
	Season     []SeasonResponse  `json:"season"`
	Procedure  string            `json:"procedure"`
	Owner      UserListItem      `json:"owner"`
	SharedUser []UserListItem    `json:"shared_user"`
	Title      string            `json:"title"`
	CreateTime *int              `json:"create_time"`
	CreateFor  int               `json:"create_for"`
	ArchiveFlg bool              `json:"archive_flg"`
	Label      []LabelResponse   `json:"label"`
}

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
	cooking := make([]CookingResponse, 0, len(r.Cooking))
	for i := range r.Cooking {
		c := &r.Cooking[i]
		cooking = append(cooking, CookingResponse{
			Ingredients: NamedRef{ID: c.Ingredient.ID, Name: c.Ingredient.Name},
			Quantity:    c.Quantity,
			Unit:        c.Unit,
		})
	}

	season := make([]SeasonResponse, 0, len(r.Season))
	for i := range r.Season {
		s := &r.Season[i]
		season = append(season, SeasonResponse{
			Seasoning: NamedRef{ID: s.Seasoning.ID, Name: s.Seasoning.Name},
			Quantity:  s.Quantity,
			Unit:      s.Unit,
		})
	}

	labels := make([]LabelResponse, 0, len(r.Labels))
	for i := range r.Labels {
		labels = append(labels, LabelResponse{ID: r.Labels[i].ID, Name: r.Labels[i].Name})
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
		CreateTime: r.CreateTime,
		CreateFor:  r.CreateFor,
		ArchiveFlg: r.ArchiveFlg,
		Label:      labels,
	}
}
