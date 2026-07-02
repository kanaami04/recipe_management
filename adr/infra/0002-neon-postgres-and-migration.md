# 0002. データベース: Neon 無料 Postgres とマイグレーション運用

- ステータス: Accepted
- 日付: 2026-07-02

## コンテキスト

[ADR-0001](0001-aws-serverless-hosting.md) のサーバレス構成で使う Postgres の置き場所が要る。
AWS 内の選択肢は最小構成でも有料(RDS t4g.micro 月 $15〜、Aurora Serverless v2 は
scale-to-zero でも復帰 15 秒前後)で、「月額ほぼ $0」の目標に合わない。
DynamoDB への書き換えは GORM ベースのリポジトリ層の全面改修になり過大。

## 決定

### 1. Neon の無料プランを使う(コード変更は接続文字列レベルのみ)

- [Neon](https://neon.com) 無料枠: 0.5GB ストレージ、アイドル後 autosuspend。
  個人のレシピ管理には十分。PostgreSQL 17 互換で GORM/pgx がそのまま動く。
- AWS 外のマネージドサービスへの依存が増えるが、標準 Postgres なので
  将来 RDS 等へ移す場合も接続文字列の差し替えで済む。

### 2. リージョンはシンガポール(ap-southeast-1)に Lambda と同居させる

Neon 無料枠は東京リージョン非対応のため、最も近い ap-southeast-1 に置き、
**Lambda も同リージョン**に置く(リクエスト毎の DB 往復レイテンシを最小化)。
ユーザー→CloudFront は日本のエッジで終端するため体感差は小さい。

### 3. Lambda からは pooled 接続、DDL は direct 接続

- Lambda の `DATABASE_URL` は **pooled 接続文字列**(ホスト名 `-pooler` 付き、PgBouncer)+
  `?sslmode=require&default_query_exec_mode=simple_protocol`。
  `simple_protocol` は transaction モードの PgBouncer で pgx のプリペアドステートメントが
  壊れる問題の Neon 公式回避策(DSN だけで済み Go コード変更なし)。
- マイグレーション(DDL)は PgBouncer 経由では不安定なため **direct 接続**で行う。
- 併せて GORM の接続プールに上限(MaxOpen 2 / MaxIdle 1 / IdleTime 5min)を設定し、
  Lambda スケールアウト時の接続枯渇を防ぐ(`api/internal/database/database.go`)。

### 4. AutoMigrate は起動時から分離する

- 環境変数 `AUTO_MIGRATE`(既定 true)を導入。ローカル開発は従来どおり起動時に
  AutoMigrate、**Lambda は false**(コールドスタート毎の DDL 実行と同時実行競合を避ける)。
- スキーマ変更時はローカルから direct DSN で `go run . -migrate` を実行する
  (マイグレーションのみ実行して終了する専用フラグ)。

## 結果

### 良い点

- DB コスト $0。コード変更は接続文字列と設定フラグに収まり、ドメイン層に影響なし。
- ローカル開発(Docker の Postgres + 起動時 AutoMigrate)は従来のまま。

### トレードオフ

- autosuspend からの復帰(1〜3 秒)+ Lambda コールドスタートが重なると、
  放置後の初回アクセスが数秒かかる。趣味用途として受容(必要なら有料プランや warmer で解消)。
- マイグレーションが「デプロイ前にローカルから手動実行」という運用手順になる
  (自動化するほどスキーマ変更が頻繁でないと判断)。
- 無料プランの制約(0.5GB・ブランチ数など)を超えたら有料化 or 移行の再検討が要る。
