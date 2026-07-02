# Recipe Management — Infra (AWS CDK)

S3 + CloudFront + Lambda(Function URL) + Supabase Postgres の最小最安構成(月額ほぼ $0)。
設計判断は [adr/infra/](../adr/infra/) を参照。

```
ブラウザ (PWA)
   │ https://xxxx.cloudfront.net
   ▼
CloudFront
   ├─ default → S3 (private, OAC) ← frontend/build/client
   │     └ CloudFront Function: 拡張子なし URI → /index.html
   └─ /api/* → Lambda Function URL (X-Origin-Verify 検証)
         └ provided.al2023/arm64 + Lambda Web Adapter → Echo (:8080)
               └ Supabase Postgres (transaction pooler, ap-northeast-1)
```

## 初回セットアップ

1. **Supabase**: [supabase.com/dashboard](https://supabase.com/dashboard) でプロジェクト作成
   (リージョンは **Northeast Asia (Tokyo)**)。Connect 画面から 2 つの接続文字列を控える:
   - **Transaction pooler**(ポート 6543)… Lambda 用
   - **Session pooler**(ポート 5432)… マイグレーション(DDL)用
     ※ direct 接続は無料枠では IPv6 のみのため、IPv4 回線からは session pooler を使う。
2. **シークレット**: `.env.example` をコピーして `infra/.env` を作成。
   ```bash
   cp .env.example .env
   # DATABASE_URL: transaction pooler + ?sslmode=require&default_query_exec_mode=simple_protocol
   # JWT_SECRET / ORIGIN_VERIFY_SECRET: openssl rand -hex 32 で生成
   ```
3. **マイグレーション**(スキーマ変更時も同じ):
   ```bash
   cd ../api
   DATABASE_URL='<session pooler の接続文字列>?sslmode=require' go run . -migrate
   ```
   DDL は transaction pooler では実行しないこと (adr/infra/0002 #3)。
4. **CDK bootstrap**(AWS アカウントに一度だけ):
   ```bash
   npm ci && npx cdk bootstrap aws://<account-id>/ap-northeast-1
   ```
5. **デプロイ**:
   ```bash
   mise run deploy   # build-lambda + build-web → cdk deploy
   ```
6. **CORS_ORIGIN の反映**(初回のみ)。出力された `RecipeStack.Url` を使って:
   ```bash
   npx cdk deploy -c corsOrigin=https://xxxx.cloudfront.net --require-approval never
   ```
   本番は同一オリジンのため CORS は実質使われないが、設定を正しておく。

## 2 回目以降のデプロイ

```bash
mise run deploy
```

## デプロイ後の確認

```bash
URL=https://xxxx.cloudfront.net
# 登録 → ログイン(Set-Cookie: refresh_token に Secure が付くこと)
curl -si -X POST "$URL/api/token/" -H 'Content-Type: application/json' \
  -H 'X-Requested-With: XMLHttpRequest' -d '{"username":"...","password":"..."}'
# Function URL 直叩きが 403 になること(出力 ApiFunctionUrl を使用)
curl -s -o /dev/null -w '%{http_code}\n' "$FUNCTION_URL/api/label/"
```

ブラウザでは: ログイン → レシピ CRUD → リロード(SPA フォールバック)→
スマホでホーム画面に追加(PWA)→ 再起動で自動ログイン、を確認する。

## 運用メモ

- Supabase 無料プランは **1 週間アクセスがないとプロジェクトが一時停止**され、
  ダッシュボードから手動で復帰させる必要がある(毎週使う前提で受容。adr/infra/0002)。
