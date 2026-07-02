# 0002. データベース: Supabase 無料 Postgres とマイグレーション運用

- ステータス: Accepted
- 日付: 2026-07-02

## コンテキスト

[ADR-0001](0001-aws-serverless-hosting.md) のサーバレス構成で使う Postgres の置き場所が要る。
AWS 内の選択肢は最小構成でも有料(RDS t4g.micro 月 $15〜、Aurora Serverless v2 は
scale-to-zero でも復帰 15 秒前後)で、「月額ほぼ $0」の目標に合わない。
DynamoDB への書き換えは GORM ベースのリポジトリ層の全面改修になり過大。

AWS 外の無料マネージド Postgres は Neon と Supabase が候補。
- **Neon**: autosuspend からの自動復帰(1〜3 秒)が魅力だが、東京リージョン非対応
  (最寄りはシンガポール)。
- **Supabase**: **東京リージョン対応**。無料プランは 1 週間非アクセスでプロジェクトが
  一時停止し手動復帰が必要だが、本アプリは毎週使う前提のため実質発生しない。

## 決定

### 1. Supabase の無料プランを使う(コード変更は接続文字列レベルのみ)

- 無料枠: DB 500MB。個人のレシピ管理には十分。標準 Postgres なので GORM/pgx がそのまま動く。
  Supabase 固有機能(Auth/Storage 等)は使わず、**素の Postgres としてだけ**使う。
- AWS 外のマネージドサービスへの依存が増えるが、標準 Postgres なので
  将来 RDS 等へ移す場合も接続文字列の差し替えで済む。

### 2. リージョンは東京(ap-northeast-1)に Lambda と同居させる

Supabase を Tokyo に作り、**Lambda も同リージョン**に置く
(リクエスト毎の DB 往復レイテンシを最小化)。

### 3. Lambda からは transaction pooler、DDL は session pooler(または direct)

- Lambda の `DATABASE_URL` は **transaction モードの pooler**(Supavisor、ポート 6543)+
  `?sslmode=require&default_query_exec_mode=simple_protocol`。
  `simple_protocol` は transaction モードの pooler で pgx のプリペアドステートメントが
  壊れる問題の回避策(DSN だけで済み Go コード変更なし)。
- マイグレーション(DDL)は transaction pooler では実行せず、**session モードの pooler
  (ポート 5432)** で行う。direct 接続は無料枠では **IPv6 のみ**のため、
  IPv4 回線からは session pooler を使う。
- 併せて GORM の接続プールに上限(MaxOpen 2 / MaxIdle 1 / IdleTime 5min)を設定し、
  Lambda スケールアウト時の接続枯渇を防ぐ(`api/internal/database/database.go`)。

### 4. AutoMigrate は起動時から分離する

- 環境変数 `AUTO_MIGRATE`(既定 true)を導入。ローカル開発は従来どおり起動時に
  AutoMigrate、**Lambda は false**(コールドスタート毎の DDL 実行と同時実行競合を避ける)。
- スキーマ変更時はローカルから session pooler DSN で `go run . -migrate` を実行する
  (マイグレーションのみ実行して終了する専用フラグ)。

## 結果

### 良い点

- DB コスト $0 で、東京リージョンに Lambda と同居できる(Neon だとシンガポールになる)。
- コード変更は接続文字列と設定フラグに収まり、ドメイン層に影響なし。
- ローカル開発(Docker の Postgres + 起動時 AutoMigrate)は従来のまま。

### トレードオフ

- **1 週間非アクセスでプロジェクトが一時停止**し、ダッシュボードからの手動復帰が要る
  (Neon のような自動復帰はない)。毎週利用する前提で受容。利用頻度が落ちたら
  Neon への乗り換えを再検討する(標準 Postgres 同士なので移行は容易)。
- マイグレーションが「デプロイ前にローカルから手動実行」という運用手順になる
  (自動化するほどスキーマ変更が頻繁でないと判断)。
- 無料プランの制約(500MB・2 プロジェクトなど)を超えたら有料化 or 移行の再検討が要る。
