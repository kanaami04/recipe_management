# 0005. ディレクトリ構成: feature ベース(コロケーション)

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

現状は種類別(type-based)構成で、`components/` `hooks/` `type/`(単数形) に機能横断のコードが
雑多に置かれている。recipes 機能のコードが `components/recipes` と `hooks` に分散し、
1 機能を触るのに複数フォルダを行き来する。空ファイル(`useRecipeData`)や命名の不統一も残る。
スケールしない構成になっている。

[ADR-0002](0002-routing-react-router-framework-mode.md) で framework mode に移行するため、
ルート層の置き方も併せて決める必要がある。

## 決定

### 1. feature ベース(コロケーション)を採用する

ドメインごとに `features/<name>/` へ api・components・schema・types をまとめ、
1 機能を 1 フォルダで完結させる。横断物は `shared/` に置く。

```
app/                          # framework mode のルート (appDirectory)
  root.tsx                    # HTMLドキュメント + Provider集約(QueryClient等)
  routes.ts                   # 型安全なルート定義
  entry.client.tsx
  routes/                     # 薄いルート層: loader/action + 描画委譲のみ
    login.tsx
    _protected.tsx            # 認証ガード(clientLoader) + Sidebarレイアウト
    _protected.recipes.tsx

  features/                   # ドメインごとにコロケーション
    auth/
      api/                    # login/refresh の query・mutation
      components/             # LoginForm
      schema/                 # zod スキーマ
      types.ts
    recipes/
      api/                    # recipes の query・mutation
      components/             # RecipeCard, RecipeForm, ...
      schema/
      types.ts

  shared/                     # 機能横断の共通物
    ui/                       # shadcn/ui コンポーネント
    components/               # 汎用コンポーネント(MessageAlertDialog, sidebar 等)
    lib/                      # axios インスタンス, queryClient, utils
    auth/                     # トークンストア + interceptors ([ADR-0004])
    api/generated/            # OpenAPI 生成物 ([ADR-0007])
    hooks/                    # use-mobile 等の汎用フック
  styles/index.css
```

### 2. ルートは薄く保つ

`routes/*` は `loader`/`action`/レイアウトのみを担い、UI ロジックは `features/*/components` に置く
([ADR-0002](0002-routing-react-router-framework-mode.md))。

### 3. 命名規約

- コンポーネントファイルは PascalCase、hooks・utils は camelCase。
- 型は feature 内に同居(`types.ts`)。グローバルな型置き場 `type/`(単数) は廃止。
- `@/` エイリアスは appDirectory 配下を指すよう維持する。

## 結果

### 良い点

- 1 機能のコードが 1 フォルダに集まり、変更時の行き来が減る。
- ルート(薄い)と feature(ロジック)の責務が分かれ、テストしやすい。
- 横断物が `shared/` に集約され、依存の向きが整理される。

### トレードオフ

- 既存ファイルの移動が広範囲に及ぶ(import の張り替えが多い)。
- 「これは feature か shared か」の線引きに判断が要る(横断利用が生じたら shared へ昇格)。
