package request

import "recipe-backend/internal/apigen"

// API 契約の型は openapi.yaml から生成する。
// 本パッケージは生成型(apigen)を再エクスポートし、ハンドラ層の import を安定させる。
// 構造体定義(json タグ・validate タグ)は生成物に一本化され、二重管理しない。
//
// refresh は Cookie から読むため body の型は持たない。
type (
	TokenRequest    = apigen.TokenRequest
	RegisterRequest = apigen.RegisterRequest
)
