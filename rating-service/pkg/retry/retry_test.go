package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRetryConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RetryConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: RetryConfig{
				MaxAttempts:   3,
				InitialDelay:  100 * time.Millisecond,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
				Jitter:        true,
			},
			wantErr: false,
		},
		{
			name: "invalid max attempts",
			config: RetryConfig{
				MaxAttempts:   0,
				InitialDelay:  100 * time.Millisecond,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
				Jitter:        true,
			},
			wantErr: true,
		},
		{
			name: "invalid initial delay",
			config: RetryConfig{
				MaxAttempts:   3,
				InitialDelay:  0,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
				Jitter:        true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWithRetry_Success(t *testing.T) {
	logger, _ := log.NewForTest()

	config := DefaultRetryConfig()
	config.MaxAttempts = 3

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := WithRetry(context.Background(), config, fn, DefaultRetryableError, logger)

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestWithRetry_AllAttemptsExhausted(t *testing.T) {
	logger, _ := log.NewForTest()

	config := DefaultRetryConfig()
	config.MaxAttempts = 2
	config.InitialDelay = 1 * time.Millisecond

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		return errors.New("persistent error")
	}

	err := WithRetry(context.Background(), config, fn, DefaultRetryableError, logger)

	assert.Error(t, err)
	assert.Equal(t, 2, callCount)
	assert.Contains(t, err.Error(), "function failed after 2 attempts")
}

func TestWithRetry_ContextCanceled(t *testing.T) {
	logger, _ := log.NewForTest()

	config := DefaultRetryConfig()
	config.MaxAttempts = 3
	config.InitialDelay = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		if callCount == 1 {
			cancel()
		}
		return errors.New("error")
	}

	err := WithRetry(ctx, config, fn, DefaultRetryableError, logger)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled during retry")
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 5*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.True(t, config.Jitter)
}
