package pg

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log/slog"
	"os"
	"webhook-delivery/internal/config"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

func MustMigrate(log *slog.Logger, cfg config.DbConfig) {
	if err := Migrate(log, cfg); err != nil {
		os.Exit(1)
	}
}

func Migrate(log *slog.Logger, cfg config.DbConfig) error {
	const fn = "app.pg.migrate.Migrate"
	log = log.With(slog.String("fn", fn))

	migration, err := migrate.New("file://migrations", GetConnectionString(cfg))
	if err != nil {
		log.Error("failed to create migration instance", sloglib.Error(err))
		return fmt.Errorf("%s: failed to create migration instance: %w", fn, err)
	}

	defer func(migration *migrate.Migrate) {
		if err, _ := migration.Close(); err != nil {
			log.Warn("failed to close migration", sloglib.Error(err))
		}
	}(migration)

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("failed to run migration", sloglib.Error(err))
		return fmt.Errorf("%s: failed to run migration: %w", fn, err)
	}

	log.Info("database is migrated successfully", slog.String("driver", "postgresql"))

	return nil
}
