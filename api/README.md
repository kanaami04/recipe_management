# Recipe Management — Backend (Go / Echo / GORM)

DRF からの刷新版バックエンド。REST API をレイヤードアーキテクチャ
（**ハンドラ層 / サービス層 / リポジトリ層**）で実装している。

## 技術スタック
- Go 1.25 + [Echo](https://echo.labstack.com/) v4
- [GORM](https://gorm.io/) + PostgreSQL 17（pgx ドライバ・CGO 不要）
- JWT 認証（golang-jwt v5、simplejwt 互換の最小実装）
- bcrypt パスワードハッシュ

## ディレクトリ構成
```
main.go        エントリポイント（DI 配線・サーバ起動）
internal/
  config       環境変数の読み込み
  database     GORM 接続・AutoMigrate
  domain       エンティティ + リポジトリ interface
  dto/request  リクエスト構造体（API 契約）
  dto/response レスポンス構造体（API 契約）
  repository   ★リポジトリ層（GORM データアクセス）
  service      ★サービス層（業務ロジック・権限・get-or-create）
  handler      ★ハンドラ層（Echo ハンドラ）
  middleware   JWT 認証ミドルウェア
  router       ルーティング + CORS
  pkg/jwt      トークン発行・検証
```
依存方向は handler → service → repository → DB。

## セットアップ & 起動
前提: [mise](https://mise.jdx.dev/) と Docker Desktop。

```bash
# 1. PostgreSQL を起動（Docker Desktop が起動している状態で）
docker compose up -d

# 2. API サーバー起動 → http://127.0.0.1:8000
#    起動時に AutoMigrate が走る
mise exec -- go run main.go            # デフォルトで .env を読み込む
mise exec -- go run main.go -env .env  # env ファイルを明示する場合
```

`.env`（未コミット）で接続情報を設定。雛形は `.env.example` を参照。

### ログ設定
構造化ログに `log/slog` を使用。`.env` で制御する。
- `LOG_LEVEL`: `debug` / `info`(既定) / `warn` / `error`
- `LOG_FORMAT`: `text`(既定・開発向け) / `json`(本番・ログ基盤向け)

リクエストログは Echo ミドルウェアから slog で出力（method/uri/status/latency、エラー時は error 付き）。

**リクエストID と SQL ログの相関**：
- 各リクエストに `X-Request-Id` を発行（クライアントが同ヘッダを送ればそれを使用）し、context 経由で下位層へ伝播。
- GORM の SQL は slog へ出力（通常クエリ=Debug / 200ms超の遅いクエリ=Warn / エラー=Error）。
- リクエストログと SQL ログの双方に `request_id` が付くため、1リクエストが発行した SQL を辿れる。
- 通常運用（`LOG_LEVEL=info`）では SQL は出力されない。SQL を見たいときは `LOG_LEVEL=debug`。

## API エンドポイント
| メソッド | パス | 認証 |
|---|---|---|
| POST | `/api/token/` | 不要 |
| POST | `/api/token/refresh/` | 不要 |
| GET | `/api/user_info/` | 必要 |
| GET | `/api/recipes/` | 必要 |
| POST | `/api/recipes/` | 必要 |
| PUT | `/api/recipes/:id/` | 必要 |
| DELETE | `/api/recipes/:id/` | 必要 |
| GET | `/api/users/` | 必要 |
| GET | `/api/label/` | 必要 |

認証は `Authorization: Bearer <access>`。

## テスト

テストピラミッドで構成。**単体／ハンドラ**は依存なしで高速、**結合**は実 Postgres
（[testcontainers-go](https://golang.testcontainers.org/)）を使う。

| 種別 | 対象 | 方式 | DB |
|---|---|---|---|
| 単体 | service / pkg(jwt) / dto | モック注入 | 不要 |
| ハンドラ | handler | `httptest` + モックサービス | 不要 |
| 結合 | repository | testcontainers で Postgres 起動 | 必要（Docker）|

結合テストは環境変数 `RUN_INTEGRATION=1` のときだけ実行する（未設定なら Skip）。
切り替えは [testutil.RequireIntegration](internal/testutil/integration.go) と
[repository の TestMain](internal/repository/setup_test.go) が担う。

```bash
# 単体 + ハンドラのみ（高速・Docker不要）
mise exec -- go test ./...

# 結合テストも含める（Docker 必須）
#   PowerShell:
$env:RUN_INTEGRATION = "1"; mise exec -- go test ./...
#   Git Bash:
RUN_INTEGRATION=1 mise exec -- go test ./...
```

> 結合テストは TestMain で Postgres コンテナを1つ起動して全テストで共有し、各テストの冒頭で
> テーブルを TRUNCATE して分離する。普段の `go test ./...` には影響しない。
