package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/right1121/railway-control-center-simulator/pkg/appctx"
)

// RecoveryMiddleware はパニックをリカバリーし、500エラーを返すミドルウェアです
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// コンテキストからロガーを取得
				ctx := appctx.FromRequest(r)
				logger := ctx.GetLogger()

				// パニックの詳細をログに記録
				logger.Error("panic recovered",
					"error", fmt.Sprintf("%v", err),
					"stack", string(debug.Stack()),
				)

				// 500 Internal Server Errorを返す
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Internal Server Error"))
				if err != nil {
					logger.Error("failed to write response",
						"error", err,
					)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
