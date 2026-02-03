# AGENTS.md - Aegisbox Development Guide

This document provides guidelines and commands for AI agents working on the aegisbox codebase.

## Project Overview

Aegisbox is an SMPP gateway and manager API service written in Go (1.23.0). It consists of two binaries:
- **smpp-gateway**: Handles SMPP connections with service providers and routes SMS to MNOs
- **manager-api**: REST API for managing the gateway (Gin-based)

Key dependencies: Gin, pgx/v5, linxGnu/gosmpp, sqlc, envconfig, slog.

## Build Commands

```bash
# Format code and tidy go.mod
make tidy

# Run quality control (vet, staticcheck, vulncheck)
make audit

# Build both binaries
make build

# Build specific binaries
make build/gateway
make build/manager-api

# Production build (Linux AMD64)
make production/build
```

## Running the Application

```bash
# Run smpp-gateway (requires DB env vars)
make run/gateway

# Run manager-api
make run/manager-api

# Live reload with air (gateway)
make run/live/gateway

# Live reload with air (manager-api)
make run/live/manager-api
```

## Testing

```bash
# Run all tests with race detector
make test

# Run tests with coverage (opens browser)
make test/cover

# Run single test
go test -v -race ./internal/sms/... -run TestSpecificFunction
```

## Docker Operations

```bash
# Build Docker images
make docker/build

# Deploy with Docker Compose
make docker/deploy

# View logs
make docker/logs

# Stop services
make docker/stop

# Full cleanup
make docker/down
```

## Code Style Guidelines

### Imports

Organize imports in three groups separated by blank lines:
1. Standard library
2. External dependencies
3. Internal packages (github.com/thrillee/aegisbox/...)

```go
import (
    "context"
    "errors"
    "fmt"
    "log/slog"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"

    "github.com/thrillee/aegisbox/internal/config"
    "github.com/thrillee/aegisbox/internal/database"
)
```

### Naming Conventions

- **Types/Exports**: PascalCase (e.g., `DefaultSender`, `Config`)
- **Variables/Functions**: camelCase (e.g., `dbQueries`, `parsePagination`)
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for config defaults
- **Interfaces**: Noun-based, optionally with -er suffix (e.g., `Sender`, `ConnectorProvider`)
- **Packages**: Simple, lowercase (e.g., `sms`, `handlers`, `config`)

### Error Handling

- Use `errors.Is` and `errors.As` for error checking
- Return errors with context using `fmt.Errorf("%w", err)`
- Log errors with structured fields: `slog.Error("message", slog.Any("error", err))`
- Handle database errors with helper functions:
  ```go
  if errors.Is(err, pgx.ErrNoRows) {
      // Handle not found
  }
  ```

### Logging

- Use `log/slog` for structured logging
- Use `slog.DebugContext`, `slog.InfoContext`, `slog.ErrorContext` with context
- Include relevant fields: `slog.String("key", value)`, `slog.Int("count", n)`
- Add source location for debug: `AddSource: logLevel <= slog.LevelDebug`

### Context Usage

- Pass context as first parameter: `func DoSomething(ctx context.Context, ...)`
- Use context for cancellation and deadlines
- Enrich context with logging helpers:
  ```go
  logCtx := logging.ContextWithMNOConnID(ctx, connID)
  ```

### Configuration

- Use `envconfig` for environment variables
- Use `godotenv` for local development (.env file)
- Define config structs with `envconfig` tags:
  ```go
  type Config struct {
      DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
      LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
  }
  ```

### Interfaces

- Define interfaces where components need abstraction
- Use compile-time interface checks:
  ```go
  var _ Sender = (*DefaultSender)(nil)
  ```
- Keep interfaces small and focused

### Database (sqlc)

- SQL queries are in `db/query/*.sql`
- Migration files in `db/migration/*.sql`
- Generated code should not be edited manually
- Run `sqlc generate` after query changes

### HTTP Handlers (Gin)

- Use `gin.Context` for request handling
- Helper functions for common patterns:
  - `parsePagination(c *gin.Context) (limit, offset int32)`
  - `parseIDParam(c *gin.Context) (int32, error)`
  - `handleGetError(c *gin.Context, logCtx context.Context, resourceName string, err error)`
- Return structured JSON responses

### Struct Tags

- Use tags for envconfig, JSON, and database operations
- Example:
  ```go
  type ManagerAPIConfig struct {
      Addr        string        `envconfig:"API_ADDR" default:":8081"`
      ReadTimeout time.Duration `envconfig:"API_READ_TIMEOUT" default:"10s"`
  }
  ```

## Database Operations

```bash
# Run migrations manually
make db/migrate
```

## Common Patterns

### Dependency Injection

Pass dependencies explicitly:
```go
func NewProcessor(deps ProcessorDependencies) *Processor {
    return &Processor{
        router:    deps.Router,
        validator: deps.Validator,
    }
}
```

### Graceful Shutdown

Use context cancellation and wait groups:
```go
appCtx, rootCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer rootCancel()

var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    server.Start(appCtx)
}()
<-appCtx.Done()
wg.Wait()
```

### Logging Context Helpers

Use context enrichment for correlated logging:
```go
func ContextWithSegmentID(ctx context.Context, segmentID int64) context.Context
func ContextWithMNOConnID(ctx context.Context, connID string) context.Context
```
