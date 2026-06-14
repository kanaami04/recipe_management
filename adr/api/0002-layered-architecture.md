# 0002. レイヤードアーキテクチャの採用

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

バックエンド(`api/`)で、HTTP の入出力・ビジネスロジック・永続化が同じ場所に混ざると、
変更の影響範囲が読めず、テストも書きにくくなる。
責務を層で分離し、テスト容易性と変更影響の局所化を図る。

## 決定

責務ごとに層を分け、**依存方向を内側(ドメイン)へ向ける**レイヤードアーキテクチャを採用する。

### レイヤーと責務

| 層 | パッケージ | 責務 |
|---|---|---|
| ハンドラ | `internal/handler` | リクエストの bind / validate、サービス呼び出し、エラー→HTTPコード変換、DTO 変換。**ビジネスロジックを持たない**。 |
| サービス | `internal/service` | ユースケース・ビジネスロジック。サービスインターフェースを公開し、`domain` のリポジトリインターフェースにのみ依存する。 |
| リポジトリ | `internal/repository` | `domain` のリポジトリインターフェースの GORM 実装。永続化の詳細を隠蔽する。 |
| ドメイン | `internal/domain` | エンティティ + リポジトリインターフェース。**最内層で他層に依存しない**。 |

補助・横断:

- `internal/dto/request`・`internal/dto/response` … API 契約(入出力 DTO)
- `internal/middleware` … JWT 認証・ログ・RequestID などの横断関心事
- `internal/pkg` … jwt など汎用ユーティリティ
- `internal/config` … 設定
- `internal/router` … ルート定義のみ
- `internal/app` … 合成ルート(Composition Root)

### 依存方向と依存性逆転

```
handler ──> service ──> domain <── repository
                          ▲
                       (interface)
```

- 外側の層が内側の層に依存する(handler → service → domain)。
- **リポジトリの抽象(interface)を `domain` に置く**。サービスは抽象にのみ依存し、
  具体実装(`repository`)は外側で注入する(依存性逆転)。
- `domain` は他のどの層にも依存しない。

### DI と合成ルート

- 各層は `NewXxx` コンストラクタで依存を受け取る(コンストラクタインジェクション)。
- `app.New` が唯一の合成ルートで、下位 → 上位に配線する
  (`repository.New` → `service.New` → `handler.New`)。
- `router.Register` はルート定義のみを担い、配線・ミドルウェア適用は `app` が行う。

### context の伝播

全層のメソッドは第一引数に `context.Context` を取り、request_id などを
下位層・GORM ログまで伝播させる。

## 結果

### 良い点

- 責務が分離され、変更の影響範囲が層内に収まりやすい。
- インターフェース越しにモックできるため、各層を独立してテストできる
  (service はリポジトリのモック、handler はサービスのモック、repository は testcontainers での結合テスト。[ADR-0001](0001-testing-aaa-and-conventions.md) 参照)。
- DB やフレームワーク(GORM / Echo)の差し替えに対する耐性が上がる。

### トレードオフ

- インターフェース・コンストラクタ・DTO 変換などのボイラープレートが増える。
- 小さな機能追加でも複数の層をまたぐ必要がある。

## 具体例(レシピ作成の流れ)

```
POST /api/recipes/
  → handler.RecipeHandler.Create   // bind/validate、userID 取り出し、DTO 変換
  → service.RecipeService.Create    // 正規化・関連解決などの業務ルール
  → domain.RecipeRepository.Create  // 抽象(interface)
  → repository.recipeRepository     // GORM 実装(app で注入)
```
