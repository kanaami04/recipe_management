# 0003. データフェッチと状態管理: TanStack Query への統一

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

現状のデータ層は中途半端に三層化している(`useApiData` → `useFetchData` →
`useRecipeData`(空ファイル))。SWR を導入しているのに、更新系(mutation)は各コンポーネントで
`api.post` / `api.delete` を直書きし、手動で `mutate("/api/recipes/")` を呼んでいる。
fetcher はトークンを引数で受け取りヘッダを手で組み立てており、一貫した方針がない。

サーバ状態(キャッシュ・再取得・楽観更新・無効化)を一貫して扱う仕組みが必要。

## 決定

### 1. サーバ状態は TanStack Query に統一する

`useApiData` / `useFetchData` / `useRecipeData` の独自三層を廃止し、
サーバ状態は **TanStack Query** に一本化する。クライアント状態(UI の開閉等)は
React の `useState` / Context のままで、サーバ状態と混同しない。

> なお 2026-05 の TanStack npm サプライチェーン攻撃(CVE-2026-45321)は **Router/Start のみ**が対象で、
> **Query は影響を受けていない**。最新版を [ADR-0009](0009-coding-conventions.md) の依存安全運用
> (lockfile 固定・`npm ci`)で使う。

### 2. query / mutation フックは OpenAPI から生成する

データ取得・更新フックは手書きせず、OpenAPI から生成する([ADR-0007](0007-api-type-sync-openapi.md))。
これにより独自フックの大半が不要になり、型と実体が契約に追従する。

### 3. 更新は mutation + 無効化に集約する

各コンポーネントでの `api.post`/`delete` 直書きと手動 `mutate` を廃止し、
mutation の成功時に対象 query key を `invalidateQueries` する。必要に応じて楽観更新を行う。

### 4. queryKey 規約

`queryKey` は feature 単位で定数化し(例: `['recipes']` / `['recipes', id]`)、文字列の直書きを避ける。

## 結果

### 良い点

- キャッシュ・再取得・楽観更新・無効化が一貫した仕組みに乗る。
- 生成フックにより手書きデータ層がほぼ消え、契約ドリフトに強くなる。
- 手動 `mutate` の貼り忘れによる「更新したのに一覧が古い」バグが構造的に減る。

### トレードオフ

- SWR から TanStack Query への置き換えと、生成パイプラインの整備コストがかかる。
- クライアント状態とサーバ状態の線引きをチームで共有する必要がある。
