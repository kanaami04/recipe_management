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

	// 接続プールの上限を控えめに固定する。
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
// 中間テーブル recipe_shares は Recipe の many2many 定義から自動生成される。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Recipe{},
		&domain.RecipeIngredient{},
		&domain.RecipeSeasoning{},
		&domain.RecipeLabel{},
		&domain.RecipeOrder{},
		&domain.RecipeArchive{},
	); err != nil {
		return err
	}
	return migrateArchiveToPerUser(db)
}

// migrateArchiveToPerUser は旧・レシピ単位の archived スカラー列を、ユーザー単位の
// recipe_archives へ一度だけ移す移行処理。AutoMigrate は列の削除・データ移動を表現できない
// ため、ここだけ生 SQL で行う。列が残っている間だけ動き、済んだら列を落とすので冪等。
// 新規 DB では archived 列自体が作られないため何もしない。
func migrateArchiveToPerUser(db *gorm.DB) error {
	var hasColumn bool
	// 接続の search_path 上の recipes(後続の ALTER/INSERT の対象)に限定する。
	// スキーマ非限定だと別スキーマの同名テーブルに誤ヒットし、DROP が失敗しうる。
	if err := db.Raw(
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = 'recipes' AND column_name = 'archived'
		)`,
	).Scan(&hasColumn).Error; err != nil {
		return err
	}
	if !hasColumn {
		return nil
	}
	// アーカイブ済みだったレシピを、その所有者のアーカイブとして移す(冪等)。移行後は
	// 列を落とし、以降の再マイグレーションで古い値がアーカイブを蘇らせないようにする。
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`INSERT INTO recipe_archives (user_id, recipe_id)
			 SELECT owner_id, id FROM recipes WHERE archived = true
			 ON CONFLICT DO NOTHING`,
		).Error; err != nil {
			return err
		}
		return tx.Exec(`ALTER TABLE recipes DROP COLUMN archived`).Error
	})
}
