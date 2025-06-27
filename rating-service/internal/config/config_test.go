package config

import (
	"os"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/notification"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
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
	assert.Equal(t, 8080, config.ServerPort)
	assert.NotEmpty(t, config.DSN)
	assert.NotNil(t, config.NotificationService)
}

func TestConfigLoadWithEnvironmentVariables(t *testing.T) {
	logger := log.New()

	_ = os.Setenv("APP_SERVER_PORT", "9080")
	_ = os.Setenv("APP_DSN", "postgres://test:test@localhost:5432/test_db")
	_ = os.Setenv("APP_NOTIFICATION_SERVICE", `{"baseUrl": "http://localhost:9081", "timeout": 30000000000}`)

	defer func() {
		_ = os.Unsetenv("APP_SERVER_PORT")
		_ = os.Unsetenv("APP_DSN")
		_ = os.Unsetenv("APP_NOTIFICATION_SERVICE")
	}()
	config, err := Load("../../config/local.yml", logger)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Check that environment variables override config values
	assert.Equal(t, 9080, config.ServerPort)
	assert.Equal(t, "postgres://test:test@localhost:5432/test_db", config.DSN)
	assert.Equal(t, "http://localhost:9081", config.NotificationService.BaseURL)
	assert.Equal(t, 30*time.Second, config.NotificationService.Timeout)

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
				ServerPort: 8080,
				DSN:        "postgres://user:pass@localhost:5432/dbname",
				NotificationService: notification.Config{
					BaseURL: "http://localhost:8081",
					Timeout: 30 * time.Second,
				},
			},
			hasErr: false,
		},
		{
			name: "missing DSN",
			config: Config{
				ServerPort: 8080,
				DSN:        "",
				NotificationService: notification.Config{
					BaseURL: "http://localhost:8081",
					Timeout: 30 * time.Second,
				},
			},
			hasErr: true,
		},
		{
			name: "invalid notification service",
			config: Config{
				ServerPort: 8080,
				DSN:        "postgres://user:pass@localhost:5432/dbname",
				NotificationService: notification.Config{
					BaseURL: "",
					Timeout: 30 * time.Second,
				},
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
