package config

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/thrillee/aegisbox/internal/workers"
)

// Config holds the overall application configuration.
type Config struct {
	DatabaseURL         string          `envconfig:"DATABASE_URL" required:"true"`
	LogLevel            string          `envconfig:"LOG_LEVEL" default:"info"`
	MNOClientConfig     MNOClientConfig // Assuming MNOClientConfig defined in smppclient pkg or here
	WorkerConfig        workers.Config  // Assuming workers.Config defined in workers pkg or here
	ServerConfig        ServerConfig
	DLRForwardInterval  time.Duration `envconfig:"WORKER_DLR_INTERVAL" default:"2s"`
	DLRForwardBatchSize int           `envconfig:"WORKER_DLR_BATCH_SIZE" default:"100"`
}

// ServerConfig holds SMPP Server specific configuration.
type ServerConfig struct {
	Addr           string        `envconfig:"SERVER_ADDR" default:":2775"`
	ReadTimeout    time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"60s"`
	WriteTimeout   time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
	BindTimeout    time.Duration `envconfig:"SERVER_BIND_TIMEOUT" default:"5s"`
	EnquireLink    time.Duration `envconfig:"SERVER_ENQUIRE_LINK" default:"30s"`
	MaxConnections int           `envconfig:"SERVER_MAX_CONNECTIONS" default:"100"`
}

// MNOClientConfig placeholder (define actual fields needed)
type MNOClientConfig struct {
	// Example:
	ConnectTimeout     time.Duration `envconfig:"MNO_CONNECT_TIMEOUT" default:"5s"`
	DefaultEnquireLink time.Duration `envconfig:"MNO_ENQUIRE_LINK" default:"30s"`
}

// workers.Config placeholder (define actual fields needed)
type WorkerConfig struct {
	RoutingInterval    time.Duration `envconfig:"WORKER_ROUTING_INTERVAL" default:"1s"`
	PricingInterval    time.Duration `envconfig:"WORKER_PRICING_INTERVAL" default:"1s"`
	SendingInterval    time.Duration `envconfig:"WORKER_SENDING_INTERVAL" default:"500ms"`
	LowBalanceInterval time.Duration `envconfig:"WORKER_LOW_BALANCE_INTERVAL" default:"5m"`
	RoutingBatchSize   int           `envconfig:"WORKER_ROUTING_BATCH_SIZE" default:"1000"`
	PricingBatchSize   int           `envconfig:"WORKER_PRICING_BATCH_SIZE" default:"1000"`
	SendingBatchSize   int           `envconfig:"WORKER_SENDING_BATCH_SIZE" default:"5000"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	log.Println("Loading configuration from environment variables...")
	err := envconfig.Process("", &cfg) // Use "" prefix for env vars
	if err != nil {
		return nil, err
	}
	log.Printf("Configuration loaded successfully (Server Addr: %s)", cfg.ServerConfig.Addr)
	return &cfg, nil
}
