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
- [0006 テーブル設計の見直し: マスタ+中間テーブル構成の廃止と命名の統一](api/0006-database-schema-redesign.md)
- [0007 ログイン識別子を username から email に変更する](api/0007-login-with-email.md)
- [0008 主キーを自動増分から UUIDv7 に変更する](api/0008-uuidv7-primary-keys.md)

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
- [0010 PWA 対応: インストール可能 + 静的アセット precache](frontend/0010-pwa-installable-precache.md)

### インフラ (infra)

- [0001 ホスティング: S3 + CloudFront + Lambda(Function URL) のサーバレス構成](infra/0001-aws-serverless-hosting.md)
- [0002 データベース: Supabase 無料 Postgres とマイグレーション運用](infra/0002-supabase-postgres-and-migration.md)
