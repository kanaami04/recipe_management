# ADR (Architecture Decision Record)

このプロジェクトの設計・運用上の意思決定を記録する。
バックエンド・フロントエンド・インフラで分けて管理する。

- [`api/`](api/) … バックエンド(Go)の ADR
- [`frontend/`](frontend/) … フロントエンド(React)の ADR
- [`infra/`](infra/) … インフラ(AWS / デプロイ)の ADR。両レイヤーにまたがる判断もここに置く

## 命名規則

`<連番4桁>-<英小文字ケバブのタイトル>.md`（例: `0001-testing-aaa-and-conventions.md`）。
連番は `api/` / `frontend/` の各ディレクトリ内で独立して採番する。

## フォーマット

各 ADR は最低限、以下を含める。

- タイトル
- ステータス（Proposed / Accepted / Deprecated / Superseded）
- 日付
- コンテキスト（なぜ決める必要があるか）
- 決定（何を決めたか）
- 結果（トレードオフ・影響）

## 一覧

### バックエンド (api)

- [0001 テストの方針: AAA パターンと検証スタイルの統一](api/0001-testing-aaa-and-conventions.md)
- [0002 レイヤードアーキテクチャの採用](api/0002-layered-architecture.md)
- [0003 エラー設計: センチネルエラーと HTTP コードへのマッピング](api/0003-error-handling.md)
- [0004 リフレッシュトークンの Cookie 化と CORS/CSRF 方針](api/0004-cookie-based-refresh-token.md)
- [0005 OpenAPI を API 契約の単一ソースとする](api/0005-openapi-contract.md)

### フロントエンド (frontend)

- [0001 ビルド基盤とアプリ形態: Vite + SPA の維持](frontend/0001-build-tooling-and-app-shape.md)
- [0002 ルーティング: React Router v7 framework mode への移行](frontend/0002-routing-react-router-framework-mode.md)
- [0003 データフェッチと状態管理: TanStack Query への統一](frontend/0003-data-fetching-tanstack-query.md)
- [0004 認証とトークン管理: メモリ保持の access + httpOnly Cookie の refresh](frontend/0004-auth-and-token-management.md)
- [0005 ディレクトリ構成: feature ベース(コロケーション)](frontend/0005-directory-structure-feature-based.md)
- [0006 フォームとバリデーション: React Hook Form + Zod](frontend/0006-forms-react-hook-form-zod.md)
- [0007 API 型とバックエンド同期: OpenAPI を契約の単一ソースとする](frontend/0007-api-type-sync-openapi.md)
- [0008 テスト方針: Vitest + Testing Library + MSW](frontend/0008-testing-vitest-rtl-msw.md)
- [0009 コーディング規約](frontend/0009-coding-conventions.md)

### インフラ (infra)

- [0001 ホスティング: S3 + CloudFront + Lambda(Function URL) のサーバレス構成](infra/0001-aws-serverless-hosting.md)
- [0002 データベース: Neon 無料 Postgres とマイグレーション運用](infra/0002-neon-postgres-and-migration.md)
