package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type ManagerAPIConfig struct {
	Addr         string        `envconfig:"API_ADDR"          default:":8081"`
	ReadTimeout  time.Duration `envconfig:"API_READ_TIMEOUT"  default:"10s"`
	WriteTimeout time.Duration `envconfig:"API_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout  time.Duration `envconfig:"API_IDLE_TIMEOUT"  default:"60s"`
	// Add Auth related config? JWT secrets?
}

// Config holds the overall application configuration.
type Config struct {
	DatabaseURL                string          `envconfig:"DATABASE_URL"              required:"true"`
	LogLevel                   string          `envconfig:"LOG_LEVEL"                                 default:"info"`
	MNOClientConfig            MNOClientConfig // Assuming MNOClientConfig defined in smppclient pkg or here
	WorkerConfig               WorkerConfig    // Assuming workers.Config defined in workers pkg or here
	ServerConfig               ServerConfig
	HttpConfig                 HttpConfig
	OtpScopePreprocessorConfig OtpScopePreprocessorConfig
	DLRForwardInterval         time.Duration `envconfig:"WORKER_DLR_INTERVAL"                       default:"2s"`
	DLRForwardBatchSize        int           `envconfig:"WORKER_DLR_BATCH_SIZE"                     default:"100"`
	DefaultRoutingGroupRef     string        `envconfig:"DEFAULT_ROUTING_GROUP_REF"                 default:"001"`
	ManagerAPI                 ManagerAPIConfig
}

// ServerConfig holds SMPP Server specific configuration.
type ServerConfig struct {
	Addr           string        `envconfig:"SERVER_HOST"            default:"0.0.0.0:2775"`
	ReadTimeout    time.Duration `envconfig:"SERVER_READ_TIMEOUT"    default:"60s"`
	WriteTimeout   time.Duration `envconfig:"SERVER_WRITE_TIMEOUT"   default:"10s"`
	BindTimeout    time.Duration `envconfig:"SERVER_BIND_TIMEOUT"    default:"5s"`
	EnquireLink    time.Duration `envconfig:"SERVER_ENQUIRE_LINK"    default:"30s"`
	MaxConnections int           `envconfig:"SERVER_MAX_CONNECTIONS" default:"100"`
}

type OtpScopePreprocessorConfig struct {
	OtpExtractionRegex string `envconfig:"SERVER_HOST"`
	DefaultAppName     string `envconfig:"SERVER_HOST" default:""`
	DefaultValidityMin string `envconfig:"SERVER_HOST" default:""`
}

// MNOClientConfig placeholder (define actual fields needed)
type MNOClientConfig struct {
	// Example:
	SyncInterval       time.Duration `envconfig:"MNO_SYNC_INTERVAL"   default:"2s"`
	ConnectTimeout     time.Duration `envconfig:"MNO_CONNECT_TIMEOUT" default:"5s"`
	DefaultEnquireLink time.Duration `envconfig:"MNO_ENQUIRE_LINK"    default:"30s"`
}

// workers.Config placeholder (define actual fields needed)
type WorkerConfig struct {
	RoutingInterval        time.Duration `envconfig:"WORKER_ROUTING_INTERVAL"          default:"1s"`
	PricingInterval        time.Duration `envconfig:"WORKER_PRICING_INTERVAL"          default:"1s"`
	SendingInterval        time.Duration `envconfig:"WORKER_SENDING_INTERVAL"          default:"500ms"`
	LowBalanceInterval     time.Duration `envconfig:"WORKER_LOW_BALANCE_INTERVAL"      default:"5m"`
	RoutingBatchSize       int           `envconfig:"WORKER_ROUTING_BATCH_SIZE"        default:"1000"`
	PricingBatchSize       int           `envconfig:"WORKER_PRICING_BATCH_SIZE"        default:"1000"`
	SendingBatchSize       int           `envconfig:"WORKER_SENDING_BATCH_SIZE"        default:"5000"`
	DLRForwardingBatchSize int           `envconfig:"WORKER_DLR_FORWARDING_BATCH_SIZE" default:"5000"`
	DLRForwarderInterval   time.Duration `envconfig:"WORKER_DLR_FORWARDING_INTERVAL"   default:"1s"`
}

type HttpConfig struct {
	Addr         string        `json:"addr"          envconfig:"HTTP_CONFIG_ADDR"          default:"0.0.0.0:8000"`
	ReadTimeout  time.Duration `json:"read_timeout"  envconfig:"HTTP_CONFIG_READ_TIMEOUT"  default:"1s"`
	WriteTimeout time.Duration `json:"write_timeout" envconfig:"HTTP_CONFIG_WRITE_TIMEOUT" default:"1s"`
	IdleTimeout  time.Duration `json:"idle_timeout"  envconfig:"HTTP_CONFIG_IDLE_TIMEOUT"  default:"1s"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	log.Println("Loading configuration from environment variables...")

	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file found, skipping: %v", err)
	} else {
		log.Println(".env loaded")
	}

	err := envconfig.Process("", &cfg) // Use "" prefix for env vars
	if err != nil {
		return nil, err
	}
	log.Printf("Configuration loaded successfully (Server Addr: %s)", cfg.ServerConfig.Addr)
	return &cfg, nil
}
