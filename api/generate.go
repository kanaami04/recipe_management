package main

// openapi.yaml を契約の単一ソースとし、Go の型を生成する (api ADR-0005)。
// 再生成: `go generate ./...`(api ディレクトリで)または `mise run gen-api`。
//
//go:generate go tool oapi-codegen -config oapi-codegen.config.yaml openapi.yaml
