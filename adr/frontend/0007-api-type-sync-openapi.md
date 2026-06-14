# 0007. API 型とバックエンド同期: OpenAPI を契約の単一ソースとする

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

API 型をフロントで手書きしており(`type/RecipeDataType.ts` 等)、Go の DTO
(`dto/response/recipe.go` 等)と契約が同期していない。片方を変えても気づけない(ドリフト)。
実際に歪みが出ている(`create_time` は型では number だが UI では string 管理、
`label`/`shared_user` の `[] | []` という不自然な型など)。

バックエンドには OpenAPI/Swagger が無い。契約を 1 箇所に固定し、ドリフトを
コンパイル時・実行時の両方で防ぐ仕組みが必要。

## 決定

### 1. OpenAPI を契約の単一ソースにする

手書きの `openapi.yaml` を契約の単一ソースとし、Go と TypeScript の双方の型をそこから生成する。
エンドポイントは少数(約 7 本)のため、手書き spec の維持は現実的。

```
openapi.yaml（手書き・単一の源）
   ├─ Go:  oapi-codegen      → request/response 型(backend 契約を固定) … api ADR-0005
   └─ FE:  生成ツール          → ① TS型  ② zodスキーマ  ③ TanStack Query フック
```

バックエンド側の OpenAPI 提供は対になる [api ADR-0005](../api/0005-openapi-contract.md) で定める。

### 2. API レスポンスの zod は「生成」する(手書きしない)

`@hey-api/openapi-ts` 等で OpenAPI から TS 型・zod バリデータ・TanStack Query フックを一括生成する。
これにより [ADR-0003](0003-data-fetching-tanstack-query.md) の手書きデータ層(`useApiData` 等)が不要になる。

### 3. コンパイル時(生成型) + 実行時(zod)の二重防御

OpenAPI は「形」の契約に過ぎず、「実際にその形で返るか」の保証は別問題。生成 zod バリデータで
API 境界のレスポンスを実行時に検証し、ドリフトを dev で即発覚させる。

### 4. 手書き zod はフォーム用に限定する

API レスポンス用 zod は生成、フォーム用 zod は手書き([ADR-0006](0006-forms-react-hook-form-zod.md))で住み分ける。
両者は形が異なる(API 形 vs UI 形)ため、`features/*/api` の変換層で繋ぐ。

### 5. 生成物の置き場所

`shared/api/generated/` に置く([ADR-0005](0005-directory-structure-feature-based.md))。
コミット対象とするか・ビルド時生成とするかは [ADR-0009](0009-coding-conventions.md) の運用に従う。

## 結果

### 良い点

- 契約が 1 ファイルに固定され、Go/TS の型が同時に追従する(ドリフト解消)。
- コンパイル時(生成型)と実行時(zod)で二重に守れる。
- 生成フックでデータ層の手書きがほぼ消える。

### トレードオフ

- OpenAPI spec を手書きで維持する手間と、生成パイプラインの整備が必要。
- バックエンド側の対応(oapi-codegen)が前提になる([api ADR-0005](../api/0005-openapi-contract.md))。
