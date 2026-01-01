package middleware

import (
	"net/http"

	"github.com/right1121/railway-control-center-simulator/pkg/appctx"
	"github.com/right1121/railway-control-center-simulator/pkg/logger"
)

// ContextMiddleware はリクエストごとに新しいcontextを生成するミドルウェアです
func ContextMiddleware(next http.Handler, baseLogger *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// トレースIDを生成
		ctx := logger.WithTraceID(r.Context())

		// リクエスト情報を生成
		reqInfo := logger.NewRequestInfo(r)

		// ロガーを設定
		reqLogger := baseLogger.WithContext(ctx).
			WithRequestInfo(ctx, reqInfo)

		// 新しいコンテキストを作成
		ctx = appctx.NewContext(ctx, reqLogger)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
