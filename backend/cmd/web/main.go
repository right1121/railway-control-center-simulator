package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/right1121/railway-control-center-simulator/internal/config"
	"github.com/right1121/railway-control-center-simulator/internal/interfaces/app"
	applogger "github.com/right1121/railway-control-center-simulator/pkg/logger"
)

func main() {
	// ロガーの初期化
	logger := applogger.New(
		applogger.WithLevel(applogger.LevelInfo),
		applogger.WithFormat("json"),
	)
	applogger.SetDefault(logger)

	// コマンドライン引数から設定ファイルのパスを取得
	configPath := flag.String("config", "", "設定ファイルのパス")
	flag.Parse()

	if *configPath == "" {
		logger.Error("--config を指定してください")
		return
	}

	ctx := context.Background()

	// 設定ファイルの読み込み
	cfg, err := config.LoadFromPath(ctx, *configPath)
	if err != nil {
		logger.Error("設定の読み込みに失敗: %v", err)
		return
	}

	// アプリケーションの初期化
	base := app.NewBase(cfg, logger)
	webApp := app.NewWebApp(base)
	go func() {
		if err := webApp.Start(); err != nil {
			logger.Error("Web app failed to start: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := webApp.Stop(ctx); err != nil {
		logger.Error("シャットダウン失敗: %v", err)
	}
}
