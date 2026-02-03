package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pressly/goose/v3"

	_ "github.com/lib/pq"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

func main() {
	mode := flag.String("mode", "dev", "Environment mode (dev, prod)")
	flag.Parse()

	var cfg Config
	log.Println("Loading configuration...")

	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file found: %v", err)
	}

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to process config: %v", err)
	}

	log.Printf("Running migrations in %s mode", *mode)

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	migrationsDir := "./db/migration"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Fatalf("Migrations directory not found: %s", migrationsDir)
	}

	log.Printf("Running migrations from: %s", migrationsDir)
	if err := goose.Up(db, migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully!")
}

func GetDBString(cfg *Config) string {
	return cfg.DatabaseURL
}
