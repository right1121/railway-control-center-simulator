package logger

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	// TraceIDKey はコンテキストに保存されるトレースIDのキー
	traceIDKey contextKey = "trace_id"
)

// NewTraceID は新しいトレースIDを生成します
func NewTraceID() string {
	u := uuid.New()
	return u.String()
}

// WithTraceID は新しいトレースIDをコンテキストに追加します
func WithTraceID(ctx context.Context) context.Context {
	return context.WithValue(ctx, traceIDKey, NewTraceID())
}

// GetTraceID はコンテキストからトレースIDを取得します
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}
