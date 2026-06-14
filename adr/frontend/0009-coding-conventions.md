# 0009. コーディング規約

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

各論点の議論で、横断的な「雑さ」が複数見つかった。`console.log` の散在、`alert()` による UX、
typo(`UserLoginFrom` / `setISArchive`)、baseURL のハードコード、URL 先頭スラッシュの混在
(`api/users/` vs `/api/recipes/`)、日英ラベルの混在、ESLint の `ecmaVersion: 2020`(tsconfig は ES2022)
とのズレ、Prettier 不在など。これらを 1 つの規約に集約する。

## 決定

### 1. 設定は環境変数で管理する

`VITE_` 接頭辞 + `import.meta.env` を使い、`.env.example` をコミットする。baseURL のハードコードは禁止。
dev は Vite proxy で `/api` 相対に寄せる([ADR-0004](0004-auth-and-token-management.md))。

### 2. console.log 禁止

ESLint `no-console`(`warn`/`error` か専用 logger のみ許容)。デバッグログは撤去する。

### 3. alert() 廃止 → トースト

shadcn の sonner を導入。確認系は `MessageAlertDialog`、通知系はトーストで使い分ける。

### 4. 命名統一

- コンポーネントファイルは PascalCase、hooks・utils は camelCase。
- typo を修正する(`LoginForm` / `setIsArchive`)。
- **ユーザー向け文言は日本語、コード識別子は英語**で統一する。

### 5. エンドポイント規約

URL の先頭スラッシュを統一する。エンドポイントは OpenAPI 生成クライアントに集約し、
手書き URL を散らさない([ADR-0007](0007-api-type-sync-openapi.md))。

### 6. エラーハンドリング

catch で握りつぶさず「ユーザー向けトースト + dev ログ」。framework mode の `ErrorBoundary` export で
ルート単位の捕捉を行う([ADR-0002](0002-routing-react-router-framework-mode.md))。

### 7. Lint / Format 強化

- Prettier を導入する。
- ESLint の `ecmaVersion` を 2022 以上に上げ、tsconfig(ES2022)と揃える。
- import order ルールと型安全系ルールを追加する。`any` は禁止し、型は `z.infer` / 生成型を優先する。

### 8. 依存の安全運用(TanStack 事件の教訓)

2026-05 の TanStack npm サプライチェーン攻撃(CVE-2026-45321)を教訓に、依存運用を明文化する。

- lockfile は必ずコミットする。
- CI は `npm ci` で固定インストールする。
- Dependabot + `npm audit` で監視する。
- 新規依存の追加は慎重に行う(本当に必要か・メンテ状況を確認)。

## 結果

### 良い点

- ログ・通知・命名・URL・依存運用の基準が 1 箇所にまとまり、レビューが容易になる。
- 環境変数化と Lint/Format 強化で、本番ノイズと表記ゆれを構造的に防げる。
- 依存安全運用を明文化し、サプライチェーン攻撃の被害を抑える。

### トレードオフ

- 既存コードの widespread な修正(ログ撤去・命名・URL 統一)が必要。
- Prettier / ESLint ルール強化で、初回は差分が大きくなる。
