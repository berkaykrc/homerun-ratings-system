package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
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
			name: "zero max attempts",
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
			name: "too many max attempts",
			config: RetryConfig{
				MaxAttempts:   15,
				InitialDelay:  100 * time.Millisecond,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
				Jitter:        true,
			},
			wantErr: true,
		},
		{
			name: "zero initial delay",
			config: RetryConfig{
				MaxAttempts:   3,
				InitialDelay:  0,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
				Jitter:        true,
			},
			wantErr: true,
		},
		{
			name: "zero max delay",
			config: RetryConfig{
				MaxAttempts:   3,
				InitialDelay:  100 * time.Millisecond,
				MaxDelay:      0,
				BackoffFactor: 2.0,
				Jitter:        true,
			},
			wantErr: true,
		},
		{
			name: "invalid backoff factor",
			config: RetryConfig{
				MaxAttempts:   3,
				InitialDelay:  100 * time.Millisecond,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 0.5,
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

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 5*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.True(t, config.Jitter)

	// Validate that default config is valid
	assert.NoError(t, config.Validate())
}

func TestDefaultRetryableError(t *testing.T) {
	err := errors.New("test error")
	assert.True(t, DefaultRetryableError(err))

	assert.True(t, DefaultRetryableError(nil))
}

func TestWithRetry_Success(t *testing.T) {
	logger, _ := log.NewForTest()
	config := DefaultRetryConfig()
	ctx := context.Background()

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		return nil // Success on first try
	}

	err := WithRetry(ctx, config, fn, nil, logger)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	logger, _ := log.NewForTest()
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Millisecond, // Very short delay for testing
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false, // No jitter for predictable testing
	}
	ctx := context.Background()

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil // Success on third try
	}

	start := time.Now()
	err := WithRetry(ctx, config, fn, nil, logger)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
	// Should have waited for at least 2 retries (1ms + 2ms)
	assert.True(t, duration >= 3*time.Millisecond)
}

func TestWithRetry_AllAttemptsExhausted(t *testing.T) {
	logger, _ := log.NewForTest()
	config := RetryConfig{
		MaxAttempts:   2,
		InitialDelay:  1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}
	ctx := context.Background()

	callCount := 0
	testErr := errors.New("persistent error")
	fn := func(ctx context.Context) error {
		callCount++
		return testErr
	}

	err := WithRetry(ctx, config, fn, nil, logger)
	assert.Error(t, err)
	assert.Equal(t, 2, callCount)
	assert.Contains(t, err.Error(), "function failed after 2 attempts")
	assert.ErrorIs(t, err, testErr)
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	logger, _ := log.NewForTest()
	config := DefaultRetryConfig()
	ctx := context.Background()

	callCount := 0
	testErr := errors.New("non-retryable error")
	fn := func(ctx context.Context) error {
		callCount++
		return testErr
	}

	isRetryable := func(err error) bool {
		return false // Mark error as non-retryable
	}

	err := WithRetry(ctx, config, fn, isRetryable, logger)
	assert.Error(t, err)
	assert.Equal(t, 1, callCount) // Should only be called once
	assert.ErrorIs(t, err, testErr)
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	logger, _ := log.NewForTest()
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond, // Long enough to be cancelled
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		return errors.New("test error")
	}

	// Cancel context after first attempt
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := WithRetry(ctx, config, fn, nil, logger)
	assert.Error(t, err)
	assert.Equal(t, 1, callCount) // Should only be called once before cancellation
	assert.Contains(t, err.Error(), "context cancelled during retry")
}

func TestCalculateDelay(t *testing.T) {
	config := RetryConfig{
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	// Test exponential backoff
	delay1 := calculateDelay(1, config)
	delay2 := calculateDelay(2, config)
	delay3 := calculateDelay(3, config)

	assert.Equal(t, 100*time.Millisecond, delay1)
	assert.Equal(t, 200*time.Millisecond, delay2)
	assert.Equal(t, 400*time.Millisecond, delay3)

	// Test max delay limit
	config.InitialDelay = 600 * time.Millisecond
	delay4 := calculateDelay(2, config)
	assert.Equal(t, 1*time.Second, delay4) // Should be capped at MaxDelay
}

func TestCalculateDelay_WithJitter(t *testing.T) {
	config := RetryConfig{
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}

	// With jitter, delays should vary slightly
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		delays[i] = calculateDelay(1, config)
	}

	// Check that at least some delays are different (jitter is working)
	hasVariation := false
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			hasVariation = true
			break
		}
	}

	// All delays should be around 100ms (base delay) but with some variation
	for _, delay := range delays {
		assert.True(t, delay >= 100*time.Millisecond)
		assert.True(t, delay <= 120*time.Millisecond) // 100ms + 10% jitter + some margin
	}

	// There should be some variation due to jitter (though this could occasionally fail due to randomness)
	assert.True(t, hasVariation, "Expected some variation in delays due to jitter")
}

func TestWithRetry_CustomRetryableError(t *testing.T) {
	logger, _ := log.NewForTest()
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}
	ctx := context.Background()

	retryableErr := errors.New("retryable error")
	nonRetryableErr := errors.New("non-retryable error")

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		if callCount == 1 {
			return retryableErr
		}
		return nonRetryableErr
	}

	isRetryable := func(err error) bool {
		return err.Error() == "retryable error"
	}

	err := WithRetry(ctx, config, fn, isRetryable, logger)
	assert.Error(t, err)
	assert.Equal(t, 2, callCount) // First call returns retryable error, second returns non-retryable
	assert.ErrorIs(t, err, nonRetryableErr)
}

func BenchmarkWithRetry_Success(b *testing.B) {
	logger, _ := log.NewForTest()
	config := DefaultRetryConfig()
	ctx := context.Background()

	fn := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WithRetry(ctx, config, fn, nil, logger)
	}
}

func BenchmarkWithRetry_WithRetries(b *testing.B) {
	logger, _ := log.NewForTest()
	config := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Microsecond, // Very short for benchmarking
		MaxDelay:      10 * time.Microsecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callCount := 0
		fn := func(ctx context.Context) error {
			callCount++
			if callCount < 3 {
				return errors.New("temp error")
			}
			return nil
		}
		_ = WithRetry(ctx, config, fn, nil, logger)
	}
}
