// Package openapi は API 契約(../openapi.yaml)からの型生成をまとめる (api ADR-0005)。
// このパッケージ自体に実装はなく、生成のエントリポイントとしてのみ存在する。
//
// 再生成: `go generate ./...`(api ディレクトリで)または `mise run gen-api`。
//
//go:generate go tool oapi-codegen -config oapi-codegen.config.yaml ../openapi.yaml
package openapi
