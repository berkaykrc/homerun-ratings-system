package config

import (
	"os"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/circuitbreaker"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoad(t *testing.T) {
	logger := log.New()

	// Test loading default config with local.yml
	config, err := Load("../../config/local.yml", logger)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Check default values
	assert.Equal(t, 8081, config.ServerPort)
	assert.NotNil(t, config.Retry)
	assert.NotNil(t, config.CircuitBreaker)
	assert.NotNil(t, config.Cleanup)
}

func TestConfigLoadWithEnvironmentVariables(t *testing.T) {
	logger := log.New()

	_ = os.Setenv("APP_SERVER_PORT", "9081")
	_ = os.Setenv("APP_RETRY", `{"maxAttempts": 5, "initialDelay": 100000000, "maxDelay": 5000000000, "backoffFactor": 2.0, "jitter": true}`)
	_ = os.Setenv("APP_CIRCUIT_BREAKER", `{"failureThreshold": 5, "recoveryTimeout": 60000000000, "minimumRequests": 3}`)
	_ = os.Setenv("APP_CLEANUP", `{"interval": 600000000000, "maxAge": 3600000000000}`)
	defer func() {
		_ = os.Unsetenv("APP_SERVER_PORT")
		_ = os.Unsetenv("APP_RETRY")
		_ = os.Unsetenv("APP_CIRCUIT_BREAKER")
		_ = os.Unsetenv("APP_CLEANUP")
	}()
	config, err := Load("../../config/local.yml", logger)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Check that environment variables override config values
	assert.Equal(t, 9081, config.ServerPort)
	assert.Equal(t, 5, config.Retry.MaxAttempts)
	assert.Equal(t, 5, config.CircuitBreaker.FailureThreshold)
	assert.Equal(t, 10*time.Minute, config.Cleanup.Interval)
	assert.Equal(t, 1*time.Hour, config.Cleanup.MaxAge)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		hasErr bool
	}{
		{
			name: "valid config",
			config: Config{
				ServerPort: 8081,
				Retry: retry.RetryConfig{
					MaxAttempts:   3,
					InitialDelay:  100 * time.Millisecond,
					MaxDelay:      5 * time.Second,
					BackoffFactor: 2.0,
					Jitter:        true,
				},
				CircuitBreaker: circuitbreaker.Config{
					FailureThreshold: 5,
					RecoveryTimeout:  60 * time.Second,
					MinimumRequests:  3,
				},
				Cleanup: CleanupConfig{
					Interval: 5 * time.Minute,
					MaxAge:   1 * time.Hour,
				},
			},
			hasErr: false,
		},
		{
			name: "invalid server port",
			config: Config{
				ServerPort: 0,
			},
			hasErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
