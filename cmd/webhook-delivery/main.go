package main

import (
	"context"
	"fmt"
	"log/slog"
	"webhook-delivery/internal/app"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/lib/logging"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

func main() {
	cfg := config.MustLoad()
	logger := logging.MustLoad(cfg.Env)

	logger.Info("logging and config were loaded successfully")

	application := app.NewApp(*cfg)

	go func() {
		if err := application.Http.Run(); err != nil {
			logger.Error("failed to listen and serve http", sloglib.Error(err))
		}
	}()

	logger.Info("http server is running", slog.Int("port", int(cfg.Port)))

	_, _ = fmt.Scanln()

	logger.Info("http server was graceful closed", slog.Int("port", int(cfg.Port)))
	if err := application.Http.Shutdown(context.Background()); err != nil {
		logger.Error("failed to shut down server", sloglib.Error(err))
	}
}
