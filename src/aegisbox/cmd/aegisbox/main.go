// Command aegisbox runs every Aegisbox backend component from one binary:
//
//	aegisbox [all]    run migrations, then the manager API and SMPP gateway together (default);
//	                  with MNOSIM_ENABLED=true the simulated MNO runs in the same process too
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
	"strconv"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
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
		if mnosimEnabled() {
			// The simulator is a test aid: if it cannot start, log it and keep the API and
			// gateway running rather than taking real traffic down with it.
			g.Go(func() error {
				if err := runEmbeddedMNOSim(gctx, cfg.DatabaseURL); err != nil {
					slog.Error("Simulated MNO stopped with an error; API and gateway keep running", slog.Any("error", err))
				}
				return nil
			})
		}
		return g.Wait()
	default:
		return fmt.Errorf("unknown command %q (use all, migrate, api, gateway or mnosim)", cmd)
	}
}

// embeddedMNOSimAddr is where the simulator listens when it runs inside "all" and MNOSIM_ADDR
// is not set: loopback only, so the gateway in the same process reaches it as 127.0.0.1:2776
// and nothing outside the container can.
const embeddedMNOSimAddr = "127.0.0.1:2776"

// mnosimEnabled reports whether MNOSIM_ENABLED asks "all" to run the simulated MNO too.
func mnosimEnabled() bool {
	v, err := strconv.ParseBool(os.Getenv("MNOSIM_ENABLED"))
	return err == nil && v
}

// loadMNOSimConfig reads the MNOSIM_* settings. defaultAddr, when not empty, replaces the listen
// address if MNOSIM_ADDR is not set.
func loadMNOSimConfig(ctx context.Context, defaultAddr string) (mnosim.Config, error) {
	cfg, err := mnosim.LoadConfig()
	if err != nil {
		return cfg, err
	}
	if defaultAddr != "" && os.Getenv("MNOSIM_ADDR") == "" {
		cfg.Addr = defaultAddr
	}
	if os.Getenv("MNOSIM_PASSWORD") == "" {
		slog.WarnContext(ctx, "MNOSIM_PASSWORD is not set; the simulated MNO uses its default password")
	}
	return cfg, nil
}

// runMNOSim runs the simulated MNO on its own (the "mnosim" command) until ctx is cancelled.
func runMNOSim(ctx context.Context) error {
	cfg, err := loadMNOSimConfig(ctx, "")
	if err != nil {
		return err
	}
	srv, err := mnosim.New(cfg)
	if err != nil {
		return err
	}
	return srv.Run(ctx)
}

// runEmbeddedMNOSim records the simulator in the database (MNO, SMPP connection, and the routes
// and prices MNOSIM_ROUTING_GROUPS asks for), then runs it until ctx is cancelled. A failed
// registration is logged and the simulator still starts, so it can be registered by hand.
func runEmbeddedMNOSim(ctx context.Context, databaseURL string) error {
	cfg, err := loadMNOSimConfig(ctx, embeddedMNOSimAddr)
	if err != nil {
		return err
	}
	srv, err := mnosim.New(cfg)
	if err != nil {
		return err
	}
	if err := registerMNOSim(ctx, databaseURL, cfg); err != nil {
		slog.ErrorContext(ctx, "Could not register the simulated MNO in the database", slog.Any("error", err))
	}
	return srv.Run(ctx)
}

func registerMNOSim(ctx context.Context, databaseURL string, cfg mnosim.Config) error {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	price := os.Getenv("MNOSIM_PRICE")
	if price == "" {
		price = "0.01"
	}
	return mnosim.Register(ctx, pool, cfg, mnosim.Registration{
		RoutingGroupRefs: mnosim.ParseRoutingGroupRefs(os.Getenv("MNOSIM_ROUTING_GROUPS")),
		Price:            price,
	})
}
