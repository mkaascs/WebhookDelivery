package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
	"webhook-delivery/internal/config"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type App struct {
	Pool *pgxpool.Pool

	log *slog.Logger
	cfg config.DbConfig
}

func New(log *slog.Logger, cfg config.DbConfig) (*App, error) {
	const fn = "app.pg.App.New"
	log = log.With(slog.String("fn", fn))

	connectionString := fmt.Sprintf("postgres://%s:%s@%s/hookrelay?sslmode=disable",
		cfg.User, cfg.Password, cfg.Addr)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectionTimeout)
	defer cancel()

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		log.Error("failed to create pgx pool", sloglib.Error(err))
		return nil, fmt.Errorf("%s: failed to create pgx pool: %w", fn, err)
	}

	return &App{
		log:  log,
		cfg:  cfg,
		Pool: pool,
	}, nil
}

func (a *App) MustConnect() {
	if err := a.Connect(); err != nil {
		os.Exit(1)
	}
}

func (a *App) Connect() error {
	const fn = "app.pg.App.Connect"
	log := a.log.With(slog.String("fn", fn))

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.ConnectionTimeout)
	defer cancel()

	if err := a.Pool.Ping(ctx); err != nil {
		log.Error("failed to ping pgx pool", sloglib.Error(err))
		return fmt.Errorf("%s: failed to ping pgx pool: %w", fn, err)
	}

	log.Info("connected to database successfully", slog.String("driver", "postgresql"))
	return nil
}

func (a *App) Close() {
	a.Pool.Close()
}
