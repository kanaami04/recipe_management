# 0005. OpenAPI を API 契約の単一ソースとする

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

現状、API の型はバックエンドの DTO(`dto/request` / `dto/response`)とフロントの手書き型に
二重で存在し、契約が同期していない(ドリフト)。バックエンドには OpenAPI/Swagger が無い。

フロントは「OpenAPI を単一ソースに、型・zod・Query フックを生成する」方針を採る
([frontend ADR-0007](../frontend/0007-api-type-sync-openapi.md))。これに対応するバックエンド側の方針を定める。

## 決定

### 1. 手書き `openapi.yaml` を契約の単一ソースにする

エンドポイントは少数(約 7 本)のため、`openapi.yaml` を手書きで維持する。これを Go/TS 双方の生成元とする。

### 2. Go の request/response 型を oapi-codegen で生成する

`openapi.yaml` から `oapi-codegen` で request/response 型を生成し、ハンドラがその型に従うようにする。
これにより実装が契約からずれた場合にコンパイル時に検出できる。

- レイヤード([ADR-0002](0002-layered-architecture.md))は維持し、生成型は handler / dto 層で用いる
  (domain/service は transport 非依存のまま)。
- エラーレスポンスの形と HTTP コードは [ADR-0003](0003-error-handling.md) のマッピングを OpenAPI 上にも反映する。

### 3. spec とコードの整合を保つ

`openapi.yaml` の更新を契約変更の起点とし、生成物の再生成を CI 等で担保する(spec とコードの乖離を防ぐ)。

## 結果

### 良い点

- 契約が 1 ファイルに固定され、Go/TS の型が同時に追従する。
- 実装が契約からずれた場合にコンパイル時へ検出が早まる。
- フロントの生成パイプライン([frontend ADR-0007](../frontend/0007-api-type-sync-openapi.md))の前提が満たされる。

### トレードオフ

- `openapi.yaml` を手書きで維持する手間がかかる(コードファーストではない)。
- 既存ハンドラを生成型に合わせる初期コストがある。
