package database

import (
	"log/slog"

	"recipe-backend/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect は PostgreSQL へ接続した *gorm.DB を返す。SQL ログは slog へ出力する。
func Connect(dsn string, logger *slog.Logger) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(logger),
	})
}

// Migrate は全エンティティの AutoMigrate を実行する。
// FK 解決のため、参照される側を先に並べる。
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.ApplicationUser{},
		&domain.RecipeLabel{},
		&domain.Ingredient{},
		&domain.Seasoning{},
		&domain.Recipe{},
		&domain.Cooking{},
		&domain.Season{},
	)
}
