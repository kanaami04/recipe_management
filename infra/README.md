# Recipe Management — Infra (AWS CDK)

S3 + CloudFront + Lambda(Function URL) + Neon Postgres の最小最安構成(月額ほぼ $0)。
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
               └ Neon Postgres (pooled DSN, ap-southeast-1)
```

## 初回セットアップ

1. **Neon**: [console.neon.tech](https://console.neon.tech) でプロジェクト作成
   (リージョンは AWS ap-southeast-1 / Singapore。東京は未対応)。
   **pooled**(ホスト名に `-pooler` 付き)と **direct** の 2 つの接続文字列を控える。
2. **シークレット**: `.env.example` をコピーして `infra/.env` を作成。
   ```bash
   cp .env.example .env
   # DATABASE_URL: pooled 接続文字列 + ?sslmode=require&default_query_exec_mode=simple_protocol
   # JWT_SECRET / ORIGIN_VERIFY_SECRET: openssl rand -hex 32 で生成
   ```
3. **マイグレーション**(スキーマ変更時も同じ):
   ```bash
   cd ../api
   DATABASE_URL='<direct の接続文字列>?sslmode=require' go run . -migrate
   ```
   DDL は pooled ではなく **direct** 接続で実行すること (adr/infra/0002 #3)。
4. **CDK bootstrap**(AWS アカウントに一度だけ):
   ```bash
   npm ci && npx cdk bootstrap aws://<account-id>/ap-southeast-1
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
