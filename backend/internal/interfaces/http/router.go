package http

import (
	"net/http"

	"github.com/right1121/railway-control-center-simulator/internal/config"
	"github.com/right1121/railway-control-center-simulator/internal/di"
	"github.com/right1121/railway-control-center-simulator/internal/interfaces/http/middleware"
	"github.com/right1121/railway-control-center-simulator/pkg/logger"
)

type Router struct {
	handler http.Handler
}

func NewRouter(cfg *config.Config, logger *logger.Logger, container *di.Container) *Router {
	mux := http.NewServeMux()

	routed := setup(mux, cfg, container)

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

func setup(mux *http.ServeMux, cfg *config.Config, container *di.Container) *http.ServeMux {
	h := NewHandler(container)

	// ヘルスチェック
	mux.Handle("GET /health", http.HandlerFunc(h.HealthCheck))

	// セッション
	mux.Handle("GET /api/session", http.HandlerFunc(h.sessionHandler.Get))
	mux.Handle("POST /api/session/join", http.HandlerFunc(h.sessionHandler.Join))
	mux.Handle("POST /api/session/leave", http.HandlerFunc(h.sessionHandler.Leave))

	return mux
}
