package requestid

import "context"

type ctxKey struct{}

// WithRequestID は context に request_id を載せる。
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext は context から request_id を取り出す（無ければ ""）。
func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}
