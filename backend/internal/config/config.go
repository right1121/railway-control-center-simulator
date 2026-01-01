package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/right1121/railway-control-center-simulator/pkg/logger"
)

type Config struct {
	Environment string `json:"environment"`
	Server      struct {
		Port int    `json:"port"`
		Host string `json:"host"`
	} `json:"server"`
	SecurePath string `json:"securePath,omitempty"`
}

func LoadFromPath(ctx context.Context, configPath string) (*Config, error) {
	logger.GetDefault().Info("load config", "path", configPath)

	config := &Config{}

	// 基本設定ファイルを読み込み
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("JSONのパースに失敗: %w", err)
	}

	return config, nil
}
