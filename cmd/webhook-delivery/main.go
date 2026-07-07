package main

import (
	"fmt"
	"webhook-delivery/internal/config"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)
}
