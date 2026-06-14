# 0008. テスト方針: Vitest + Testing Library + MSW

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

バックエンドはテスト方針が確立している(AAA + testify + 「〜の時、〜こと。」コメント。
[api ADR-0001](../api/0001-testing-aaa-and-conventions.md))。一方フロントエンドのテストは
ゼロで、語彙・構造の基準もない。一貫性のため、バックエンドの哲学を踏襲しつつ
フロント向けのツールを定める。

## 決定

### 1. ツール

| 役割 | 採用 | 理由 |
|---|---|---|
| テストランナー | Vitest | Vite ネイティブで設定流用・高速・ESM |
| コンポーネント | React Testing Library + user-event | 実装でなく振る舞いを検証する標準 |
| API モック | MSW (Mock Service Worker) | ネットワーク層で intercept。dev とテストで共用でき、axios interceptor もそのまま通る |
| マッチャ | @testing-library/jest-dom | DOM アサーションを読みやすく |

### 2. バックエンド ADR-0001 を踏襲する規約

- **構造は AAA**: `// Arrange` `// Act` `// Assert` で区切る。
- **意図コメント**: テスト直前に「〜の時、〜こと。」形式で 1 文書く。
- **1 テスト 1 検証**: 1 つの振る舞いを 1 テストで検証する。
- **セットアップ集約 + ファクトリ**: テストデータは functional options 風のファクトリへ
  (backend の `testutil/factory` に対応)。`renderWithProviders`(QueryClient / Router / Context をまとめる)
  ヘルパーを用意する。

### 3. 重点を置く対象(費用対効果、カバレッジ 100% は追わない)

- **zod スキーマ**(フォーム検証): 必須・数値変換・件数の境界。
- **form → API 変換関数**([ADR-0006](0006-forms-react-hook-form-zod.md))。
- **認証 interceptor / refresh の single-flight**([ADR-0004](0004-auth-and-token-management.md)):
  401→refresh→リトライ、同時多発で 1 回集約、失敗でログアウト。バグりやすいため厚く。
- **clientLoader の認証ガード**: 未ログインで `redirect`。
- **主要コンポーネント**: LoginForm 送信、RecipeForm の作成/編集を RTL + MSW で。

### 4. E2E は将来導入(任意)

ログイン → 一覧 → 作成 の主要フローを Playwright で 1〜2 本。MVP では必須にせず、将来導入とする。

## 結果

### 良い点

- バックエンドと語彙・構造が揃い、レビューと差分把握が容易になる。
- ロジックの濃い箇所(検証・変換・認証)に投資が集中する。
- MSW でネットワーク層をモックでき、interceptor を含めた現実に近い検証ができる。

### トレードオフ

- テスト基盤(Vitest 設定・MSW ハンドラ・ヘルパー)の初期整備コストがかかる。
- カバレッジを追わない方針のため、何をテストするかの判断をチームで共有する必要がある。
