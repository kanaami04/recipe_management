package database

import (
	"log/slog"
	"time"

	"recipe-backend/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect は PostgreSQL へ接続した *gorm.DB を返す。SQL ログは slog へ出力する。
func Connect(dsn string, logger *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(logger),
	})
	if err != nil {
		return nil, err
	}

	// 接続プールの上限を控えめに固定する (adr/infra/0002)。
	// Lambda は 1 インスタンス 1 リクエストの低並列で、無制限のままだと
	// スケールアウト時に DB の接続数を食い潰すため。ローカルでも支障はない。
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
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
