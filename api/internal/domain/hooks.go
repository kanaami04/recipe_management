package domain

import (
	"recipe-backend/internal/pkg/id"

	"gorm.io/gorm"
)

// GORM の BeforeCreate フックで、未採番(空文字)のときだけ UUIDv7 を採番する。
// ID を明示指定して作るケース(テスト等)ではその値を尊重する。
//
// domain が gorm / pkg/id に依存するのは、エンティティが既に gorm タグへ密結合している
// 延長として許容する(採番ロジックをエンティティ自身に持たせる方が凝集度が高い)。

func assignID(dst *string) error {
	if *dst == "" {
		*dst = id.New()
	}
	return nil
}

func (u *User) BeforeCreate(*gorm.DB) error             { return assignID(&u.ID) }
func (r *Recipe) BeforeCreate(*gorm.DB) error           { return assignID(&r.ID) }
func (i *RecipeIngredient) BeforeCreate(*gorm.DB) error { return assignID(&i.ID) }
func (s *RecipeSeasoning) BeforeCreate(*gorm.DB) error  { return assignID(&s.ID) }
func (l *RecipeLabel) BeforeCreate(*gorm.DB) error      { return assignID(&l.ID) }
func (l *Label) BeforeCreate(*gorm.DB) error            { return assignID(&l.ID) }
func (l *ShoppingList) BeforeCreate(*gorm.DB) error     { return assignID(&l.ID) }
func (i *ShoppingListItem) BeforeCreate(*gorm.DB) error { return assignID(&i.ID) }
