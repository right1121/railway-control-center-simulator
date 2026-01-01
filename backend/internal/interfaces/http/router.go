package http

import (
	"net/http"

	"github.com/right1121/railway-control-center-simulator/internal/config"
	"github.com/right1121/railway-control-center-simulator/internal/interfaces/http/middleware"
	"github.com/right1121/railway-control-center-simulator/pkg/logger"
)

type Router struct {
	handler http.Handler
}

func NewRouter(cfg *config.Config, logger *logger.Logger) *Router {
	mux := http.NewServeMux()

	routed := setup(mux, cfg)

	// ミドルウェアを順番に適用
	handler := middleware.ContextMiddleware(
		middleware.RecoveryMiddleware(routed),
		logger,
	)

	return &Router{
		handler: handler,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func setup(mux *http.ServeMux, cfg *config.Config) *http.ServeMux {
	h := NewHandler()

	// ヘルスチェック
	mux.Handle("GET /health", http.HandlerFunc(h.HealthCheck))

	return mux
}
