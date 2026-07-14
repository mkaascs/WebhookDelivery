package main

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
	"webhook-delivery/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up/down>")
	}

	cfg := config.MustLoad()

	cmd := os.Args[1]
	connectionString := fmt.Sprintf("postgres://%s:%s@%s/hookrelay?sslmode=disable",
		cfg.User, cfg.Password, cfg.DbConfig.Addr)

	migration, err := migrate.New("file://migrations", connectionString)
	if err != nil {
		log.Fatalf("failed to create migration instance: %v", err)
	}

	defer func(migration *migrate.Migrate) {
		if err, _ := migration.Close(); err != nil {
			log.Fatalf("failed to close migration: %v", err)
		}
	}(migration)

	cmds := map[string]func() error{
		"up":   migration.Up,
		"down": func() error { return migration.Steps(-1) },
	}

	fn, ok := cmds[cmd]
	if !ok {
		log.Fatalf("unknown command: %s", cmd)
	}

	if err := fn(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to migrate db: %v", err)
	}

	log.Printf("migrated successfully")
}
