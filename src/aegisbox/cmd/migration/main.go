package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/thrillee/aegisbox/internal/app"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file found: %v", err)
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to process config: %v", err)
	}

	if err := app.RunMigrations(context.Background(), cfg.DatabaseURL); err != nil {
		log.Fatalf("%v", err)
	}
	log.Println("Migrations completed successfully!")
}
