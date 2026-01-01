package http

import (
	"net/http"
)

type Handler struct{}

func NewHandler(container *di.Container) *Handler {
}

// HealthCheck はヘルスチェックを行うハンドラです
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		panic(err)
	}
}
