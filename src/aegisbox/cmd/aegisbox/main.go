// Command aegisbox runs every Aegisbox backend component from one binary:
//
//	aegisbox [all]    run migrations, then the manager API and SMPP gateway together (default)
//	aegisbox migrate  run database migrations and exit
//	aegisbox api      run only the manager API
//	aegisbox gateway  run only the SMPP gateway
//	aegisbox mnosim   run the simulated MNO (test traffic only; configured with MNOSIM_* variables)
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/thrillee/aegisbox/internal/app"
	"github.com/thrillee/aegisbox/internal/config"
	"github.com/thrillee/aegisbox/internal/mnosim"
)

func main() {
	cmd := "all"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// The simulated MNO needs no database or gateway configuration.
	if cmd == "mnosim" {
		app.SetupLogging(os.Getenv("LOG_LEVEL"))
		if err := runMNOSim(ctx); err != nil {
			slog.Error("Simulated MNO stopped with an error", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	app.SetupLogging(cfg.LogLevel)

	if err := run(ctx, cmd, cfg); err != nil {
		slog.Error("Aegisbox stopped with an error", slog.String("command", cmd), slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd string, cfg *config.Config) error {
	switch cmd {
	case "migrate":
		return app.RunMigrations(ctx, cfg.DatabaseURL)
	case "api":
		return app.RunManagerAPI(ctx, cfg)
	case "gateway":
		return app.RunGateway(ctx, cfg)
	case "all":
		if err := app.RunMigrations(ctx, cfg.DatabaseURL); err != nil {
			return err
		}
		// If either component fails, the group context cancels the other so the
		// process exits and the container restarts both together.
		g, gctx := errgroup.WithContext(ctx)
		g.Go(func() error { return app.RunManagerAPI(gctx, cfg) })
		g.Go(func() error { return app.RunGateway(gctx, cfg) })
		return g.Wait()
	default:
		return fmt.Errorf("unknown command %q (use all, migrate, api, gateway or mnosim)", cmd)
	}
}

func runMNOSim(ctx context.Context) error {
	cfg, err := mnosim.LoadConfig()
	if err != nil {
		return err
	}
	srv, err := mnosim.New(cfg)
	if err != nil {
		return err
	}
	return srv.Run(ctx)
}
