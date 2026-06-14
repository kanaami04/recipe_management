# 0002. ルーティング: React Router v7 framework mode への移行

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

現状は React Router v7 を **declarative(library) mode** で使い、`<Routes><Route>` を
JSX で並べている([現 `AppRouter.tsx`])。ここに 2 つの問題がある。

1. **認証ガードが壊れている**: `let isAuthenticated = false` がモジュールレベルの
   グローバル可変変数で、一度 `true` になるとログアウトしても戻らない。
2. **描画してから取得**: 各ページがコンポーネント内で fetch するため、
   画面を出してからデータ取得が始まる。

React Router v7 は Remix を取り込み、**framework mode**(`loader`/`action`・型安全ルート)を
本命として打ち出している。SPA を維持したまま(`ssr: false`)その恩恵を受けられる。

本アプリは [ADR-0001](0001-build-tooling-and-app-shape.md) で SPA 維持を決めている。また個人開発であり、
移行コスト(工数・学習)を理由に避ける必要が薄い。規模が小さい今が移行の好機でもある。

## 決定

### 1. React Router v7 を framework mode へ移行する(`ssr: false`)

`@react-router/dev` の Vite プラグインを使い、framework mode の **SPA 構成**(`ssr: false`)で運用する。
静的配信・Node サーバ不要という [ADR-0001](0001-build-tooling-and-app-shape.md) の前提は維持される。

### 2. ルートは薄く保ち、データは loader で先読みする

- ルートモジュールは `loader` / `clientLoader` / `action` / レイアウトのみを担い、
  UI ロジックは feature 側へ委譲する([ADR-0005](0005-directory-structure-feature-based.md))。
- データは `clientLoader` で遷移前に先読みする。実体の取得・キャッシュは
  TanStack Query に委ねる([ADR-0003](0003-data-fetching-tanstack-query.md))。

### 3. 認証ガードは clientLoader に集約する

保護ルートの `clientLoader` で未ログインを判定し `redirect("/")` する。
現状のグローバル可変変数による判定は廃止する。アクセストークンの参照方法は
[ADR-0004](0004-auth-and-token-management.md) に従う(loader は React の外側で動くため、
モジュールレベルのストアから読む)。

### 4. 型安全なルートを使う

ルート定義から型を生成し、パラメータ・URL を型付きで扱う。

## 結果

### 良い点

- `loader` 先読みで「描画してから取得」を解消できる。
- 認証ガードがルート定義に集約され、グローバル可変状態のバグが消える。
- 型安全ルートで URL/パラメータの取り違えをコンパイル時に防げる。
- 業界標準のメンタルモデル(loader/action)を学べる。

### トレードオフ

- エントリ・ルーティング・データ取得の土台を同時に作り替えるため、段階移行が効きにくい。
- 「コンポーネントで fetch」→「loader で先読み」へのメンタルモデルの切り替えコストがある。

### 将来

画面数が増えてルーティングが複雑化した場合や、一部に SSR が必要になった場合は、
`ssr: true` への切り替えを別 ADR で再検討する。
