# 0008. 主キーを自動増分から UUIDv7 に変更する

- ステータス: Accepted
- 日付: 2026-07-02

## コンテキスト

全エンティティの主キーは GORM の autoIncrement(bigserial)だった。連番 ID は
以下の弱みがある。

- **推測可能**: URL の `/api/recipes/1/` から他レコードの存在・件数が推測でき、
  列挙的なアクセスを誘発する。
- **採番が DB 依存**: ID を確定するには INSERT が必要で、アプリ側で事前に採番できない。
- **分散・マージに弱い**: 別環境のデータ統合や事前生成が難しい。

一方、完全ランダムな UUIDv4 は B-Tree インデックスの挿入位置がばらつき、断片化して
書き込みが劣化する。**UUIDv7** は先頭にミリ秒精度のタイムスタンプを持ち、時刻順に
概ね昇順で採番されるため、連番に近い挿入局所性を保ちつつ推測困難性を得られる。

## 決定

### 1. 全エンティティの主キー・FK を UUIDv7(Postgres の uuid 型)にする

- Go 側の型は `string`、Postgres 側は `uuid` 型(`gorm:"type:uuid;primaryKey"`)。
  ドメインを uuid ライブラリの型に縛らず、JWT / JSON / GORM いずれとも素直に相互運用できる
  string を採用する。
- `OwnerID` / `RecipeID` などの FK、共有中間テーブル `recipe_shares` の複合キーも uuid にする。

### 2. 採番用の共通関数を新設し、domain の GORM フックで採番する

- `internal/pkg/id` に `id.New() string` を置き、`github.com/google/uuid` の
  `uuid.NewV7()` で採番する(乱数源の失敗は継続不能なので `Must` で panic)。
- ドメインの各エンティティに GORM の `BeforeCreate` フックを実装し、ID が空のときだけ
  `id.New()` で採番する(`internal/domain/hooks.go`)。ID を明示指定した場合はそれを尊重する。
- **domain が gorm / pkg/id に依存する**点は ADR-0002(domain は最内層)から一歩外れるが、
  エンティティは既に gorm タグへ密結合しており、「採番ロジックをエンティティ自身に持たせる」
  方が凝集度が高いと判断し、その延長として許容する。

### 3. JWT の subject も string に揃える

- `jwt.Manager` の `GenerateAccess/GenerateRefresh/Parse` と `user_id` クレーム、
  ミドルウェアの `UserIDFromContext` を string 化する。

### 4. API 契約とパスパラメータを uuid 型にする

- `UserInfoResponse.id` / `UserListItem.id` / `RecipeResponse.id` を
  `type: string, format: uuid`(`x-go-type: string`)に、`/api/recipes/{id}` の
  パスパラメータも `format: uuid` にする。Go(oapi-codegen)/ TS(openapi-ts)双方の型を再生成。
- ハンドラの `parseID` は `uuid.Parse` で形式検証し、不正な id は 400 で弾く
  (任意文字列がそのまま WHERE 句に渡るのを防ぐ)。

### 5. DB 拡張は不要

- ID はアプリ側で採番するため、`uuid-ossp` / `pgcrypto` などの DB 拡張や
  `DEFAULT gen_random_uuid()` は使わない。uuid 型カラムに文字列を書き込むだけ。

## 結果

### 良い点

- ID が推測困難になり、件数や存在の推測・列挙が難しくなった。
- アプリ側で採番するため INSERT 前に ID が確定でき、テストやログも扱いやすい。
- UUIDv7 の時刻順序性で、v4 のようなインデックス断片化を避けられる。

### トレードオフ

- 主キーが 8 バイト整数から 16 バイト uuid になり、インデックス/ストレージがやや増える
  (個人用途の規模では無視できる)。
- bigserial 前提の既存データは型変換できないため、スキーマは作り直し
  (旧テーブルの DROP + AutoMigrate)。本アプリは本番未投入・ローカルは使い捨てのため受容。
- domain が gorm / pkg/id に依存する(上記 2 の通り、密結合の延長として許容)。
