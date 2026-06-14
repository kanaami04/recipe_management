# 0003. エラー設計: センチネルエラーと HTTP コードへのマッピング

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

エラーは複数の層をまたいで伝播する(repository → service → handler)。
このとき、

- ドメイン/サービス層が HTTP ステータスコードを知ってしまうと、層の責務が崩れる([ADR-0002](0002-layered-architecture.md))。
- エラーの種類ごとに正しい HTTP コード(401 / 403 / 404 / 409 / 400 / 500)を返したい。
- 内部実装の詳細をレスポンスに漏らしたくない。

エラーの表現方法と、HTTP コードへ変換する場所の方針を定める。

## 決定

### 1. センチネルエラーを service 層に集約する

業務的に意味のあるエラーは `internal/service/errors.go` に**センチネルエラー**として定義する。

```go
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrNotFound           = errors.New("recipe not found")
	ErrForbidden          = errors.New("forbidden")
	ErrSharedUserNotFound = errors.New("shared user not found")
)
```

### 2. 文脈はラップで付加し、判定は `errors.Is` で行う

詳細を足したい場合は `fmt.Errorf("%w: ...", ErrXxx, detail)` でラップする。
判定は文字列比較ではなく `errors.Is` を使い、ラップされていても一致させる。

```go
// service 側
return fmt.Errorf("%w: %s", ErrSharedUserNotFound, su.Username)
// handler 側
case errors.Is(err, service.ErrSharedUserNotFound): ...
```

### 3. HTTP コードへの変換は handler の責務

ドメイン/サービスは HTTP を知らない。**handler だけ**がセンチネルエラーを
HTTP ステータスへ変換する。レシピ系は共通関数 `mapServiceError` に集約する。

```go
func mapServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "...")
	case errors.Is(err, service.ErrSharedUserNotFound):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}
```

認証系は 401 / 409 が必要なため、`Token` / `Refresh` / `Register` の各ハンドラ内で個別に変換する。

### 4. エラー → HTTP コード対応表

| エラー / 状況 | HTTP | 変換場所 |
|---|---|---|
| `ErrInvalidCredentials`(ログイン) | 401 | `AuthHandler.Token` |
| refresh トークンが無効 | 401 | `AuthHandler.Refresh` |
| `ErrUserAlreadyExists` | 409 | `AuthHandler.Register` |
| `ErrNotFound` | 404 | `mapServiceError` |
| `ErrForbidden` | 403 | `mapServiceError` |
| `ErrSharedUserNotFound` | 400 | `mapServiceError` |
| bind / validate 失敗 | 400 | handler(`bindRecipe` 等、サービス呼び出し前) |
| 上記以外(想定外) | 500 | 各 handler の `default` |

### 5. 想定外エラーは 500 + 汎用メッセージ

マッピングされていないエラーは 500 とし、レスポンスには `"internal error"` のような
汎用メッセージのみを返す(内部詳細は露出しない)。詳細はログ([ADR-0002](0002-layered-architecture.md) の RequestLogger)で追う。

## 結果

### 良い点

- HTTP の知識が handler に閉じ、service/domain は transport 非依存でテストしやすい。
- `errors.Is` ベースなのでラップに強く、文脈(ユーザー名等)を足しても判定が壊れない。
- マッピングが一覧化され、どのエラーが何番になるか追いやすい。
- 内部詳細を 500 で隠蔽でき、情報漏洩を防げる。

### トレードオフ

- 新しいセンチネルエラーを追加したら、handler のマッピングにも追記が必要
  (漏れると `default` で 500 になる)。
- 認証系(個別変換)とレシピ系(`mapServiceError`)で変換箇所が分かれる
  → 401 / 409 が認証固有のため、現状は許容する。

### テスト

handler テストでセンチネルエラーを返すモックを差し込み、404 / 403 / 400 / 401 / 409 / 500 を
網羅的に検証する([ADR-0001](0001-testing-aaa-and-conventions.md))。
