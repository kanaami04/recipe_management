package repository

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"recipe-backend/internal/database"
	"recipe-backend/internal/domain"
	"recipe-backend/internal/pkg/invite"
	"recipe-backend/internal/testutil"
	"recipe-backend/internal/testutil/factory"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"
)

// testDB は結合テストで共有する DB ハンドル（TestMain で初期化）。
var testDB *gorm.DB

// TestMain は RUN_INTEGRATION=1 のときだけ Postgres コンテナを1つ起動し、
// 全結合テストで共有する。未設定なら各テストは RequireIntegration で Skip される。
func TestMain(m *testing.M) {
	if os.Getenv(testutil.EnvRunIntegration) != "1" {
		os.Exit(m.Run())
	}

	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:17",
		tcpostgres.WithDatabase("recipe"),
		tcpostgres.WithUsername("recipe"),
		tcpostgres.WithPassword("recipe"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("connection string: %v", err)
	}

	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := database.Connect(dsn, discard)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	testDB = db

	code := m.Run()

	_ = container.Terminate(ctx)
	os.Exit(code)
}

// truncateAll は全テーブルをクリアする（各テストの冒頭で呼んで分離する）。
func truncateAll(t *testing.T) {
	t.Helper()
	const stmt = `TRUNCATE recipes, recipe_labels, recipe_ingredients, recipe_seasonings,
		recipe_orders, recipe_archives, labels,
		shopping_lists, shopping_list_items,
		share_groups, share_group_members, users RESTART IDENTITY CASCADE`
	if err := testDB.Exec(stmt).Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

// seedUser はテスト用ユーザーを作成する。データ生成はファクトリに集約している。
func seedUser(t *testing.T, username string) *domain.User {
	t.Helper()
	u := factory.NewUser(factory.WithUsername(username), factory.WithEmail(username+"@example.com"))
	if err := testDB.Create(u).Error; err != nil {
		t.Fatalf("seed user %q: %v", username, err)
	}
	return u
}

// seedShareGroup は owner を所有者、members を追加メンバーとするシェアグループを作成する。
// これにより owner と members は互いの共有物を見られるようになる。
func seedShareGroup(t *testing.T, owner *domain.User, members ...*domain.User) *domain.ShareGroup {
	t.Helper()
	repo := NewShareGroupRepository(testDB)
	code, err := invite.Code()
	if err != nil {
		t.Fatalf("invite code: %v", err)
	}
	g := &domain.ShareGroup{
		Name:                "テストグループ",
		OwnerID:             owner.ID,
		InviteCode:          code,
		InviteCodeExpiresAt: time.Now().Add(time.Hour),
	}
	if err := repo.Create(context.Background(), g); err != nil {
		t.Fatalf("seed group: %v", err)
	}
	for _, m := range members {
		if err := repo.AddMember(context.Background(), g.ID, m.ID, true); err != nil {
			t.Fatalf("add member %q: %v", m.Username, err)
		}
	}
	return g
}
