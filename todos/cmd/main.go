package main

import (
	"log"
	"github.com/chienha0903/Todo_App/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Starting %s on port %s in %s mode", cfg.AppName, cfg.AppPort, cfg.AppEnv)
}