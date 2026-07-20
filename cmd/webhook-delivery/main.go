package main

//	@title			Webhook Delivery
//	@version		1.0.0
//	@description	An open-source webhook delivery service in Go

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
	cfg := config.MustLoad()
	logger := logging.MustLoad(cfg.Env)

	logger.Info("webhook-delivery is starting", slog.String("env", cfg.Env))

	application := app.New(logger, *cfg)

	application.Postgres.MustConnect()
	application.Workers.Start()

	application.MountMiddlewares()
	application.MountHandlers()

	go application.HTTP.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	application.Postgres.Close()
	application.Workers.Shutdown()
	if err := application.HTTP.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("webhook-delivery is stopped")
}
