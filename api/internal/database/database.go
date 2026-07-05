package database

import (
	"log/slog"
	"sort"
	"time"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/pkg/invite"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Recipe{},
		&domain.RecipeIngredient{},
		&domain.RecipeSeasoning{},
		&domain.RecipeLabel{},
		&domain.RecipeOrder{},
		&domain.RecipeArchive{},
		&domain.Label{},
		&domain.ShoppingList{},
		&domain.ShoppingListItem{},
		&domain.ShareGroup{},
		&domain.ShareGroupMember{},
	); err != nil {
		return err
	}
	if err := migrateArchiveToPerUser(db); err != nil {
		return err
	}
	if err := migrateSharesToGroups(db); err != nil {
		return err
	}
	return seedLabelsFromRecipes(db)
}

// tableExists は current_schema 上に name テーブルが存在するかを返す。
func tableExists(db *gorm.DB, name string) (bool, error) {
	var exists bool
	err := db.Raw(
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = current_schema() AND table_name = ?
		)`, name,
	).Scan(&exists).Error
	return exists, err
}

// migrateSharesToGroups は旧・個別共有(recipe_shares / shopping_list_shares)を
// シェアグループへ一度だけ移す移行処理。共有関係を無向グラフの辺と見なし、連結成分ごとに
// 1 グループを作って全員をメンバーにする。済んだら旧テーブルを落とすので冪等。新規 DB では
// 旧テーブルが無いため何もしない。
func migrateSharesToGroups(db *gorm.DB) error {
	hasRecipeShares, err := tableExists(db, "recipe_shares")
	if err != nil {
		return err
	}
	hasListShares, err := tableExists(db, "shopping_list_shares")
	if err != nil {
		return err
	}
	if !hasRecipeShares && !hasListShares {
		return nil
	}

	type edge struct {
		OwnerID string
		UserID  string
	}
	var edges []edge
	if hasRecipeShares {
		var e []edge
		if err := db.Raw(
			`SELECT r.owner_id AS owner_id, rs.user_id AS user_id
			 FROM recipe_shares rs JOIN recipes r ON r.id = rs.recipe_id`,
		).Scan(&e).Error; err != nil {
			return err
		}
		edges = append(edges, e...)
	}
	if hasListShares {
		var e []edge
		if err := db.Raw(
			`SELECT sl.owner_id AS owner_id, sls.user_id AS user_id
			 FROM shopping_list_shares sls JOIN shopping_lists sl ON sl.id = sls.shopping_list_id`,
		).Scan(&e).Error; err != nil {
			return err
		}
		edges = append(edges, e...)
	}

	// union-find で共有関係の連結成分を求める。
	parent := map[string]string{}
	var find func(string) string
	find = func(x string) string {
		if parent[x] == "" {
			parent[x] = x
		}
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}
	for _, e := range edges {
		if e.OwnerID == "" || e.UserID == "" || e.OwnerID == e.UserID {
			continue
		}
		union(e.OwnerID, e.UserID)
	}
	components := map[string][]string{}
	for u := range parent {
		root := find(u)
		components[root] = append(components[root], u)
	}

	// 共有されていた買い物リストの所有者集合。グループの買い物リストは「グループ所有者の
	// リスト」で解決するため、所有者にこの中の誰かを選ばないと、共有されていたリストが
	// 見えなくなる。所有者選定でこの集合を優先する。
	listShareOwners := map[string]struct{}{}
	if hasListShares {
		var owners []string
		if err := db.Raw(
			`SELECT DISTINCT sl.owner_id
			 FROM shopping_list_shares sls JOIN shopping_lists sl ON sl.id = sls.shopping_list_id`,
		).Scan(&owners).Error; err != nil {
			return err
		}
		for _, o := range owners {
			listShareOwners[o] = struct{}{}
		}
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, members := range components {
			if len(members) < 2 {
				continue
			}
			sort.Strings(members) // 決定的にするため。
			// 共有されていた買い物リストの所有者がいればそれを所有者にする(共有リストを保つ)。
			// いなければ決定的に先頭を選ぶ。
			ownerID := members[0]
			for _, m := range members {
				if _, ok := listShareOwners[m]; ok {
					ownerID = m
					break
				}
			}
			code, err := invite.Code()
			if err != nil {
				return err
			}
			group := domain.ShareGroup{
				Name:                "マイグループ",
				OwnerID:             ownerID,
				InviteCode:          code,
				InviteCodeExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}
			if err := tx.Omit("Owner", "Members").Create(&group).Error; err != nil {
				return err
			}
			for _, uid := range members {
				m := domain.ShareGroupMember{GroupID: group.ID, UserID: uid}
				if err := tx.Omit("Group", "User").
					Clauses(clause.OnConflict{DoNothing: true}).
					Create(&m).Error; err != nil {
					return err
				}
			}
		}
		// 移行済み。旧・個別共有テーブルを落とす(以降の再マイグレーションは何もしない)。
		if err := tx.Exec(`DROP TABLE IF EXISTS recipe_shares`).Error; err != nil {
			return err
		}
		return tx.Exec(`DROP TABLE IF EXISTS shopping_list_shares`).Error
	})
}

// seedLabelsFromRecipes は既存レシピに付いていたラベル名を、その所有者の Label マスタへ
// 一度だけ取り込む。所有者ごとに DISTINCT 名を作り、既にあるものは飛ばす(冪等)。
// マスタ導入前から使っていたラベルを、管理画面で扱えるようにするための移行。
// ID は他エンティティと同じ UUIDv7 にするため、生 INSERT ではなく BeforeCreate 経由で作る。
func seedLabelsFromRecipes(db *gorm.DB) error {
	type pair struct {
		OwnerID string
		Name    string
	}
	var pairs []pair
	if err := db.Raw(
		`SELECT DISTINCT r.owner_id AS owner_id, rl.name AS name
		 FROM recipe_labels rl
		 JOIN recipes r ON r.id = rl.recipe_id`,
	).Scan(&pairs).Error; err != nil {
		return err
	}
	for _, p := range pairs {
		label := domain.Label{Name: p.Name, OwnerID: p.OwnerID}
		// 既にある (owner, name) は飛ばす(冪等)。belongs-to の Owner は巻き込まない。
		if err := db.Omit("Owner").
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&label).Error; err != nil {
			return err
		}
	}
	return nil
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
