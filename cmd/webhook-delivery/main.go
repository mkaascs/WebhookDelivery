package main

import (
	"context"
	"fmt"
	"log"
	"webhook-delivery/internal/app"
	"webhook-delivery/internal/config"
)

func main() {
	cfg := config.MustLoad()

	application := app.NewApp(*cfg)

	go func() {
		if err := application.Http.Run(); err != nil {
			log.Println(err)
		}
	}()

	_, _ = fmt.Scanln()

	if err := application.Http.Shutdown(context.Background()); err != nil {
		log.Println(err)
	}
}
