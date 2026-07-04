package response

import (
	"recipe-backend/internal/apigen"
	"recipe-backend/internal/domain"
)

// LabelItem は管理対象ラベル(マスタ)の DTO。構造体定義は openapi.yaml から生成する。
type LabelItem = apigen.LabelItem

// ToLabelItem は domain.Label を API 契約の DTO へ変換する。
func ToLabelItem(l *domain.Label) LabelItem {
	return LabelItem{Id: l.ID, Name: l.Name}
}
