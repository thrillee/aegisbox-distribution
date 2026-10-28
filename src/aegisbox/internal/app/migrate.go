package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/thrillee/aegisbox/db/migration"
)

// RunMigrations applies all pending migrations embedded in the binary.
func RunMigrations(ctx context.Context, databaseURL string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	goose.SetBaseFS(migration.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	slog.Info("Running database migrations...")
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to read migration version: %w", err)
	}
	slog.Info("Database migrations finished", slog.Int64("version", version))
	return nil
}
