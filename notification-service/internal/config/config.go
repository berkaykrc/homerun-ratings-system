package config

import (
	"os"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/circuitbreaker"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/retry"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/qiangxue/go-env"
	"gopkg.in/yaml.v2"
)

const (
	defaultServerPort = 8081
)

// Config represents an application configuration.
type Config struct {
	// the server port. Defaults to 8081
	ServerPort int `yaml:"server_port" env:"SERVER_PORT"`

	// retry configuration for resilience
	Retry retry.RetryConfig `yaml:"retry" env:"RETRY"`

	// circuit breaker configuration
	CircuitBreaker circuitbreaker.Config `yaml:"circuit_breaker" env:"CIRCUIT_BREAKER"`

	// notification cleanup configuration
	Cleanup CleanupConfig `yaml:"cleanup" env:"CLEANUP"`
}

// CleanupConfig represents notification cleanup configuration
type CleanupConfig struct {
	Interval time.Duration `yaml:"interval" json:"interval"`
	MaxAge   time.Duration `yaml:"max_age" json:"maxAge"`
}

// Validate validates the cleanup configuration
func (c CleanupConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Interval, validation.Required, validation.Min(time.Minute)),
		validation.Field(&c.MaxAge, validation.Required, validation.Min(time.Minute)),
	)
}

// Validate validates the application configuration.
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ServerPort, validation.Required, validation.Min(1), validation.Max(65535)),
		validation.Field(&c.Retry, validation.Required),
		validation.Field(&c.CircuitBreaker, validation.Required),
		validation.Field(&c.Cleanup, validation.Required),
	)
}

// Load returns an application configuration which is populated from the given configuration file and environment variables.
func Load(file string, logger log.Logger) (*Config, error) {
	// default config
	c := Config{
		ServerPort:     defaultServerPort,
		Retry:          retry.DefaultRetryConfig(),
		CircuitBreaker: circuitbreaker.DefaultConfig(),
		Cleanup: CleanupConfig{
			Interval: 5 * time.Minute,
			MaxAge:   1 * time.Hour,
		},
	}

	// load from YAML config file
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(bytes, &c); err != nil {
		logger.Errorf("failed to unmarshal configuration file %s: %s", file, err)
		return nil, err
	}

	// load from environment variables prefixed with "APP_"
	if err = env.New("APP_", logger.Infof).Load(&c); err != nil {
		return nil, err
	}

	// validation
	if err = c.Validate(); err != nil {
		return nil, err
	}

	return &c, err
}
