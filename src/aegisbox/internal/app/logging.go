package app

import (
	"log/slog"
	"os"

	"github.com/thrillee/aegisbox/internal/logging"
)

// SetupLogging installs the JSON slog logger used by every Aegisbox component and returns it.
func SetupLogging(logLevelStr string) *slog.Logger {
	logLevel := slog.LevelInfo
	if logLevelStr == "debug" {
		logLevel = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logLevel <= slog.LevelDebug,
	} // Add source location for debug
	baseHandler := slog.NewJSONHandler(os.Stdout, opts)
	contextHandler := logging.NewContextHandler(baseHandler)
	logger := slog.New(contextHandler)
	slog.SetDefault(logger)
	slog.Info("Logging initialized", "level", logLevel.String())
	return logger
}
