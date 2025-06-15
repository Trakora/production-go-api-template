package config

import (
	"fmt"
	"os"
	"production-go-api-template/pkg/constants"
	"time"

	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
)

type Conf struct {
	Server   ConfServer
	Auth     ConfAuth
	Security ConfSecurity
	DB       ConfDB
}

type ConfServer struct {
	Port         int           `env:"SERVER_PORT,default=8080"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,default=30s"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,default=30s"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,default=60s"`
	Debug        bool          `env:"SERVER_DEBUG,default=true"`
	CorsOrigins  []string      `env:"SERVER_CORS_ORIGINS,default=*"`
}

type ConfAuth struct {
	APITokens   string `env:"API_TOKEN,required"`
	HMACSecrets string `env:"SECRET,required"`
}

type ConfSecurity struct {
	MaxFailures   int           `env:"SECURITY_MAX_FAILURES,default=5"`
	FailWindow    time.Duration `env:"SECURITY_FAIL_WINDOW,default=1m"`
	BlockDuration time.Duration `env:"SECURITY_BLOCK_DURATION,default=10m"`
	CleanupTick   time.Duration `env:"SECURITY_CLEANUP_TICK,default=5m"`
	SlowdownStep  time.Duration `env:"SECURITY_SLOWDOWN_STEP,default=200ms"`
	SlowdownMax   time.Duration `env:"SECURITY_SLOWDOWN_MAX,default=2s"`
}

type ConfDB struct {
	DBPath string `env:"DB_PATH,default=database.db"`
	Debug  bool   `env:"SERVER_DEBUG,default=true"`
}

const (
	defaultDotenv = ".env"
)

func New() (*Conf, error) {
	dotenvPath := os.Getenv("DOTENV_CONFIG_PATH")
	if dotenvPath == constants.EmptyString {
		dotenvPath = defaultDotenv
	}

	if err := godotenv.Load(dotenvPath); err != nil {
		return nil, fmt.Errorf("dotenv load failed: %w", err)
	}

	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &c, nil
}
