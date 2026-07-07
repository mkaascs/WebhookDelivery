package logging

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"webhook-delivery/internal/config"
)

func MustLoad(env string) *slog.Logger {
	logger, err := Load(env)
	if err != nil {
		log.Fatal(err)
	}

	return logger
}

func Load(env string) (*slog.Logger, error) {
	var logger *slog.Logger

	switch env {
	case config.EnvLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case config.EnvDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case config.EnvProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return nil, fmt.Errorf("unknown environment type: %s", env)
	}

	return logger, nil
}
