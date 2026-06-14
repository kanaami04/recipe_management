# 0001. テストの方針: AAA パターンと検証スタイルの統一

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

バックエンド(`api/`)のテストが、標準ライブラリ(`if ... t.Error`)と testify で混在し、
検証の粒度・コメント・構造がファイルごとにバラバラだった。
読みやすさ・一貫性・レビューのしやすさのために、テストの書き方を統一する。

## 決定

以下をバックエンドテストの共通方針とする。

### 1. 構造は AAA(Arrange / Act / Assert)

各テストを `// Arrange` `// Act` `// Assert` のコメントで区切る。
Arrange と Act がヘルパー内で一体になっている場合は `// Arrange & Act` と表記する。

### 2. アサーションは testify に統一

- 前提条件・nil 回避などの**ガードは `require`**(失敗したら即停止)。
- **本来の検証は `assert`**(失敗しても続行し、差分を出す)。

### 3. 1 テスト 1 検証

1 つの振る舞いを 1 テストで検証する。複数の観点を同じテストに詰め込まない。
ただし「結果に到達するための」`require` ガード(`NoError` / `NotNil` / `Len` など)は許容する。

### 4. 構造体で検証できるものは構造体ごと比較

`assert.Equal` で構造体・スライスを丸ごと比較し、全項目を一度に検証する。
タイムスタンプ(`CreatedAt` 等)や自動採番のように**非決定的な値を含む構造体は、
該当する項目だけ**を比較する(例: GORM が読み込む `ApplicationUser.DateJoined`)。

### 5. テストの意図をコメントで書く

テスト関数の直前に「〜の時、〜こと。」形式で検証内容を 1 文で書く。
(AAA コメントとは別に、関数の目的を表す。)

### 6. セットアップは集約する

- 重複するセットアップは `t.Helper()` 付きヘルパーへ。
- テストデータ生成は functional options のファクトリ(`internal/testutil/factory`)へ。
- HTTP ハンドラは `serveXxx` ヘルパー + モックサービス + `httptest` で組む。

### 7. 結合テストの扱い

- `testcontainers` で Postgres を起動し、`RUN_INTEGRATION=1` のときだけ実行する
  (未設定なら `RequireIntegration` で Skip)。
- 各テスト冒頭で `truncateAll`(`RESTART IDENTITY`)を呼び、テスト間を分離する。

### 8. ハンドラの異常系を網羅

ステータスコードとエラーマッピング(404 / 403 / 400 / 500)を検証する。
サービスが呼ばれてはいけない経路(バリデーション失敗・不正 ID 等)は、
モックに `t.Fatal` を仕込んで「呼ばれないこと」を担保する。

## 結果

### 良い点

- テストの構造・語彙が揃い、レビューと差分把握が容易になる。
- 「〜の時、〜こと。」コメントでテストの意図が一目で分かる。
- 構造体ごとの比較で、項目の検証漏れを防げる。

### トレードオフ

- 1 検証への分割でテスト数とセットアップ呼び出しが増える(ヘルパー/ファクトリ集約で緩和)。
- 構造体比較は非決定値(時刻・採番)に注意が必要。含む場合は項目比較へ切り替える。

## 具体例

```go
// 関連付きレシピを作成した時、FindByID で食材(Cooking)が構造体ごと読み込まれること。
func TestRecipeRepo_FindByID_LoadsCooking(t *testing.T) {
	// Arrange
	ctx, repo, id := arrangeRecipeWithRelations(t)

	// Act
	got := mustFindRecipe(t, ctx, repo, id)

	// Assert
	assert.Equal(t, []domain.Cooking{{
		ID: 1, RecipeID: 1, IngredientID: 1,
		Ingredient: domain.Ingredient{ID: 1, Name: "じゃがいも"},
		Quantity:   3, Unit: "個",
	}}, got.Cooking)
}
```
