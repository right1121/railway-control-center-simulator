package app

import (
	"github.com/right1121/railway-control-center-simulator/internal/config"
	"github.com/right1121/railway-control-center-simulator/pkg/logger"
)

type BaseApp struct {
	config *config.Config
	logger *logger.Logger
}

func NewBase(cfg *config.Config, log *logger.Logger) *BaseApp {
	return &BaseApp{
		config: cfg,
		logger: log,
	}
}
