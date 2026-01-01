package http

import (
	"net/http"

	"github.com/right1121/railway-control-center-simulator/internal/di"
	session "github.com/right1121/railway-control-center-simulator/internal/interfaces/http/session_handler"
)

type Handler struct {
	sessionHandler *session.SessionHandler
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{
		sessionHandler: session.NewSessionHandler(container.UseCases.Session),
	}
}

// HealthCheck はヘルスチェックを行うハンドラです
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		panic(err)
	}
}
