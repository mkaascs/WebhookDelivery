package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webhook-delivery/internal/app"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/lib/logging"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := config.MustLoad()
	logger := logging.MustLoad(cfg.Env)

	logger.Info("webhook-delivery is starting", slog.String("env", cfg.Env))

	application := app.New(logger, *cfg)

	application.Postgres.MustConnect()
	go application.Http.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	application.Postgres.Close()
	if err := application.Http.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("webhook-delivery is stopped")
}
