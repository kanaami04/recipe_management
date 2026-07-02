# 0006. テーブル設計の見直し: マスタ+中間テーブル構成の廃止と命名の統一

- ステータス: Accepted
- 日付: 2026-07-02

## コンテキスト

初期スキーマは Django 時代の設計を引き継いでおり、一般的な設計原則から見て問題が多かった。

- **過剰な正規化**: 食材・調味料・ラベルがそれぞれ「マスタテーブル + 中間テーブル」
  (`ingredients`+`cooking`、`seasoning`+`season`、`recipe_labels`+`recipes_label`)で表現されていた。
  マスタは name の get-or-create で無限に増え、参照されなくなっても掃除されない。
  食材や調味料はユーザーが自由入力する文字列であり、全ユーザー横断のマスタとして
  共有する要件もないため、正規化の利益がない。
- **Django の遺物**: テーブル名 `application_user`、未使用カラム `is_staff` / `is_superuser`、
  `date_joined` などフレームワーク都合の名残が残っていた。
- **誤解を招く命名**: `create_time`(実体は調理時間)、`create_for`(何人前)、`archive_flg`、
  ハッシュを格納する `password` カラムなど。
- **命名の不統一**: テーブル名の単複が混在(`recipes` vs `cooking` / `season`)。
- **FK 制約の欠如**: 中間テーブル・子テーブルに ON DELETE CASCADE がなく、
  リポジトリ層が手動で子行を削除していた。

## 決定

### 1. レシピ従属データは子テーブルに非正規化し、マスタ+中間テーブルをやめる

| 旧 | 新 |
|---|---|
| `ingredients` + `cooking` | `recipe_ingredients` (recipe_id, name, quantity, unit) |
| `seasoning` + `season` | `recipe_seasonings` (recipe_id, name, quantity, unit) |
| `recipe_labels` + `recipes_label` | `recipe_labels` (recipe_id, name) |

- いずれも `(recipe_id, name)` に一意制約、`recipe_id` に ON DELETE CASCADE の FK を張る。
- ラベル一覧 API は `recipe_labels.name` の DISTINCT で返す。その際、自分が閲覧できる
  (所有 or 共有された)レシピのラベルに限定する(旧実装は全ユーザーのラベルを返していた)。
- get-or-create 系のリポジトリメソッドは不要になり削除。

### 2. レシピ共有のみ中間テーブルとして残す

レシピ⇔ユーザーの共有は真の多対多であり中間テーブルが必然。`recipes_shared_user` を
`recipe_shares` (recipe_id, user_id 複合 PK、両方向 FK CASCADE) に改める。

### 3. 命名を一般的な規約に揃える

- テーブル名は複数形スネークケースに統一: `users`(`user` は PostgreSQL の予約語のため
  複数形を採用)、`recipes`、`recipe_ingredients`、`recipe_seasonings`、`recipe_labels`、`recipe_shares`。
- カラム名は実体を表す名前に: `create_time`→`cooking_time`、`create_for`→`servings`、
  `archive_flg`→`archived`、`password`→`password_hash`、`date_joined`→`created_at`。
- 未使用の `is_staff` / `is_superuser` は削除。

### 4. API 契約は原則維持し、実体を失った id のみ削除する

`create_time` / `create_for` / `archive_flg` / `cooking` / `season` などの API フィールド名は
契約の安定を優先してそのまま(改名は別途判断)。ただしラベル・食材・調味料の `id` は
マスタ廃止で意味を失ったため、`LabelResponse` と `NamedRef`(→`NameResponse`)から削除した。
フロントエンドは name しか参照していなかったため影響は型の再生成のみ。

### 5. 移行はスキーマ作り直しで行う

本番 DB は未投入・ローカルは使い捨てのテストデータのみのため、データ移行は行わず
AutoMigrate による新規作成とする。旧テーブル(`application_user`、`cooking`、`ingredients`、
`season`、`seasoning`、`recipes_label`、`recipes_shared_user`)が残っている環境では手動で
DROP する。

## 結果

### 良い点

- テーブル 9 → 6。中間テーブルは必然性のある `recipe_shares` の 1 つだけになった。
- レシピの作成・更新・削除がシンプルになった(get-or-create 廃止、削除は FK CASCADE 任せ)。
- 名前を見れば実体がわかる(cooking_time、servings、password_hash)。
- ラベル一覧が自分に関係するものだけになり、他ユーザーのラベル名が漏れなくなった。

### トレードオフ

- 同名の食材・ラベルがレシピごとに行として重複する(非正規化)。個人用途のデータ量では
  問題にならず、集計・オートコンプリートは DISTINCT で足りる。
- 食材名の一括リネーム(マスタ更新で全レシピに反映)はできなくなったが、そのような
  要件は存在しない。
- スキーマ変更に伴い既存環境では旧テーブルの手動 DROP が必要。
