# Recipe Management — Infra (AWS CDK)

S3 + CloudFront + Lambda(Function URL) + Supabase Postgres の最小最安構成(月額ほぼ $0)。

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
   DDL は transaction pooler では実行しないこと。
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

`mise run deploy` は **マイグレーション → ビルド → cdk deploy** の順で走る。
`[tasks.deploy]` が `migrate` を依存に含むため、`migrate`(スキーマ適用)は必ず
`cdk deploy`(コード稼働)より前に完了する。列を追加した新コードが列の無い DB に対して
先に動き出す事故(既存レコードの読み取り・INSERT が「列が無い」で全て 500)を、
手動の手順ではなくデプロイの構造として防ぐ。

`migrate` は `api/.env.migrate`(Session pooler の DATABASE_URL、gitignore 済み)を読む。
このファイルが無い環境ではサイレントにスキップせず、明確なエラーで `deploy` 全体を中断する。
DDL を流す端末には事前に `api/.env.migrate` を用意しておくこと(中身は初回セットアップの
session pooler 接続文字列 + `?sslmode=require`)。

> **補足**: AutoMigrate は追加系のみ・冪等のため、スキーマ変更が無いデプロイで `migrate` を
> 流しても列は追加されず実質何もしない(安全に毎回走らせてよい)。破壊的 DDL(列削除・
> 型変更・DROP)は AutoMigrate では行われない。
>
> なお本番(Lambda, `AUTO_MIGRATE=false`)は起動時に「期待カラムの有無」を検査し、
> スキーマがコードより古い場合は CloudWatch ログに
> `database schema is behind code; run mise run migrate` を出す(API は落とさない)。
> 万一 migrate を流し忘れても、ログで早期に気づける。

## スマホからデプロイ(GitHub Actions)

ローカル環境なしで、スマホの GitHub アプリからデプロイできる。仕組みは
`.github/workflows/deploy.yml`:

- **手動実行のみ**(`workflow_dispatch`)。GitHub アプリ → Actions → **Deploy** → Run workflow。
- **AWS 認証は OIDC**。静的アクセスキーはどこにも保存せず、実行のたびに数分で失効する
  一時クレデンシャルを引き受ける。
- **production Environment の承認ゲート**で、押し間違いによる誤デプロイを防ぐ。
- 中身は `mise run deploy` を呼ぶだけなので、**migrate → build → cdk deploy** の順序と
  「スキーマ適用がコード稼働より前」は 2 回目以降のデプロイと同じ。

### 初回セットアップ(一度きり)

1. **AWS に GitHub OIDC プロバイダを作成**(未作成なら)。
   - プロバイダ URL: `https://token.actions.githubusercontent.com`、対象(audience): `sts.amazonaws.com`。

2. **デプロイ用 IAM ロールを作成**。信頼ポリシーをこのリポジトリの production Environment に限定する:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [{
       "Effect": "Allow",
       "Principal": { "Federated": "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com" },
       "Action": "sts:AssumeRoleWithWebIdentity",
       "Condition": {
         "StringEquals": {
           "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
           "token.actions.githubusercontent.com:sub": "repo:kanaami04/recipe_management:environment:production"
         }
       }
     }]
   }
   ```
   権限は CDK v2 のブートストラップロールを引き受けられれば足りる(最小):
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [{
       "Effect": "Allow",
       "Action": "sts:AssumeRole",
       "Resource": "arn:aws:iam::<ACCOUNT_ID>:role/cdk-hnb659fds-*"
     }]
   }
   ```
   個人アカウントで簡単に済ませるなら、代わりに広めの管理ポリシーを付けてもよい。

3. **GitHub の Environment `production` を作成**し、**Required reviewers に自分を追加**(承認ゲート)。
   Settings → Environments → New environment。

4. **Secrets を production Environment に登録**(Settings → Environments → production → Secrets):

   | Secret | 中身 |
   |---|---|
   | `AWS_DEPLOY_ROLE_ARN` | 手順 2 で作った IAM ロールの ARN |
   | `DATABASE_URL` | **Transaction pooler**(ポート 6543)+ `?sslmode=require&default_query_exec_mode=simple_protocol`。Lambda 用 |
   | `MIGRATE_DATABASE_URL` | **Session pooler**(ポート 5432)+ `?sslmode=require`。マイグレーション(DDL)用 |
   | `JWT_SECRET` | `openssl rand -hex 32` |
   | `ORIGIN_VERIFY_SECRET` | `openssl rand -hex 32` |

   > `DATABASE_URL` と `MIGRATE_DATABASE_URL` のプーラーを取り違えないこと。DDL は transaction
   > pooler では実行できないため、migrate は必ず session pooler を使う。ワークフローは前者を
   > `infra/.env`、後者を `api/.env.migrate` へ書き分ける。

5. `npx cdk bootstrap` が未実施なら、初回セットアップ手順で一度だけ実行しておく。

### 使い方(スマホ)

GitHub アプリ → 対象リポジトリ → Actions → **Deploy** → Run workflow →
実行後に承認(Review deployments → Approve)。以降は自動で migrate → build → deploy が走る。

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
  ダッシュボードから手動で復帰させる必要がある(毎週使う前提で受容)。
