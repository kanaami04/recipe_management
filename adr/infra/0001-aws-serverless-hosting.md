# 0001. ホスティング: S3 + CloudFront + Lambda(Function URL) のサーバレス構成

- ステータス: Accepted
- 日付: 2026-07-02

## コンテキスト

個人用アプリを月額ほぼ $0 で常時公開したい。フロントは静的配信可能な SPA
([frontend ADR-0001](../frontend/0001-build-tooling-and-app-shape.md))、API は Go の常駐 HTTP サーバ。
認証は「同一オリジン + SameSite=Lax Cookie」前提([api ADR-0004](../api/0004-cookie-based-refresh-token.md))
のため、フロントと API を別オリジンに分けると Cookie/CSRF 設計が崩れる。

## 決定

### 1. S3 + CloudFront + Lambda で、CloudFront をルーターにして同一オリジンを維持する

- SPA ビルド(`frontend/build/client`)は private な S3 バケット(OAC 経由)から配信する。
- API は Lambda + **Function URL**(API Gateway は使わない。恒久無料枠の Lambda に対し
  API Gateway は $1/100万リクエストかかり、機能も今回は不要)。
- CloudFront の behavior で `/api/*` → Lambda、それ以外 → S3 に振り分ける。
  **フロントも API も同一オリジン**になり、Cookie(SameSite=Lax・Domain 未指定)と
  CSRF 方針・フロントの相対 `/api` 前提がすべて無改修で成立する。

### 2. Lambda は provided.al2023/arm64 + Lambda Web Adapter(コード無改変)

- Go バイナリを `bootstrap` 名でビルドし、AWS 公式の
  [Lambda Web Adapter](https://github.com/awslabs/aws-lambda-web-adapter) レイヤーを添える。
  LWA が Lambda イベント ↔ HTTP を変換するため、**Echo アプリは常駐サーバのまま無改変**。
  ローカルと本番が完全に同じコードパスで動く。
- 対案の aws-lambda-go-api-proxy(echoadapter)は main にイベント変換の分岐が入り、
  ローカルと別経路になるため見送り。プロキシ 1 hop 分のオーバーヘッドは許容する。

### 3. CloudFront → Lambda の保護はシークレットヘッダ検証(OAC + IAM は不採用)

- Function URL は `AuthType: NONE` とし、CloudFront が origin custom header
  `X-Origin-Verify`(ランダム値)を付与、アプリのミドルウェアが検証する
  (`api/internal/middleware/originverify.go`)。直叩きは 403。
- **OAC + IAM 認証を不採用にした理由**: OAC は origin リクエストの `Authorization`
  ヘッダを SigV4 署名で上書きするため本アプリの `Authorization: Bearer` 認証と衝突し、
  さらに POST/PUT にクライアント側で body の SHA-256(`x-amz-content-sha256`)付与が必須になる。
  回避にはフロント・バックエンド両方の改修が要り、将来 multipart(画像)を足す際の障害にもなる。

### 4. SPA フォールバックは CloudFront Function、キャッシュはパスで分ける

- viewer-request の CloudFront Function で「拡張子なし URI → /index.html」に書き換える。
  **custom error response を使わない**のは、distribution 全体に効いて `/api/*` の 403/404
  レスポンスまで index.html に化けるため。
- キャッシュ: ハッシュ付き `assets/*` は `immutable` 長期、`index.html`/`sw.js`/manifest は
  `no-cache`(ここを長期にすると Service Worker の更新が届かなくなる)。`/api/*` は CachingDisabled。
- `/api/*` の origin request policy は `AllViewerExceptHostHeader`
  (Host を転送すると Function URL のルーティングが壊れる)。

### 5. IaC は AWS CDK(TypeScript)単一スタック、シークレットは Lambda 環境変数

- `infra/` に単一スタック。ビルド(Go / フロント)は mise タスクで分離し、`mise run deploy` で一括実行。
- `DATABASE_URL` / `JWT_SECRET` / `ORIGIN_VERIFY_SECRET` は `infra/.env`(gitignore)から
  Lambda 環境変数へ注入する。SSM SecureString は CloudFormation の dynamic reference が
  Lambda 環境変数に非対応で、ランタイム取得はコード変更が要るため見送り(1 人運用の受容リスク)。
- 本番 URL は CloudFront 既定ドメイン。独自ドメインは将来 ACM(us-east-1)+ CNAME で後付け可能。

## 結果

### 良い点

- 月額ほぼ $0(Lambda / CloudFront の恒久無料枠内。S3 数円)。
- 同一オリジン維持により、認証・CSRF・API クライアントのコード変更が不要。
- アプリはローカルと同一の常駐サーバとして動き、Lambda 固有のコードを持たない。

### トレードオフ

- Lambda のコールドスタートで放置後の初回アクセスが数秒かかる
  ([ADR-0002](0002-supabase-postgres-and-migration.md))。趣味用途として受容。
- Function URL がネットワーク上は公開される(保護はアプリ層の検証のみ)。
  URL はランダムで推測困難、かつデータは JWT 認証の背後にあり、受容範囲とする。
- シークレットが Lambda 環境変数に平文で載る(コンソール閲覧可)。1 人運用のため受容。
