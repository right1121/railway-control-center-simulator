package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/right1121/railway-control-center-simulator/internal/di"
	apphttp "github.com/right1121/railway-control-center-simulator/internal/interfaces/http"
)

type WebApp struct {
	*BaseApp
	Server *http.Server
}

func NewWebApp(base *BaseApp) *WebApp {
	container := di.NewContainer(base.config)

	router := apphttp.NewRouter(base.config, base.logger, container)
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", base.config.Server.Host, base.config.Server.Port),
		Handler: router,
	}
	return &WebApp{
		BaseApp: base,
		Server:  server,
	}
}

func (a *WebApp) Start() error {
	log.Printf("Starting web server on http://%s", a.Server.Addr)
	return a.Server.ListenAndServe()
}

func (a *WebApp) Stop(ctx context.Context) error {
	log.Println("Shutting down web server...")
	return a.Server.Shutdown(ctx)
}
