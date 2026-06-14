package request

// RecipeRequest は POST/PUT /api/recipes/ のリクエスト。
type RecipeRequest struct {
	Title      string            `json:"title" validate:"required"`
	CreateTime *int              `json:"create_time"`
	CreateFor  int               `json:"create_for"`
	Procedure  string            `json:"procedure"`
	ArchiveFlg bool              `json:"archive_flg"`
	Label      []LabelInput      `json:"label"`
	SharedUser []SharedUserInput `json:"shared_user"`
	Cooking    []CookingInput    `json:"cooking"`
	Season     []SeasonInput     `json:"season"`
}

type LabelInput struct {
	Name string `json:"name" validate:"required"`
}

type SharedUserInput struct {
	Username string `json:"username" validate:"required"`
}

type NameInput struct {
	Name string `json:"name" validate:"required"`
}

type CookingInput struct {
	Ingredients NameInput `json:"ingredients"`
	Quantity    int       `json:"quantity"`
	Unit        string    `json:"unit"`
}

type SeasonInput struct {
	Seasoning NameInput `json:"seasoning"`
	Quantity  int       `json:"quantity"`
	Unit      string    `json:"unit"`
}
