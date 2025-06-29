package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_NewCircuitBreaker(t *testing.T) {
	logger, _ := log.NewForTest()
	config := DefaultConfig()

	cb := New(config, logger)

	assert.NotNil(t, cb)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	logger, _ := log.NewForTest()
	config := Config{
		FailureThreshold: 3,
		RecoveryTimeout:  time.Second,
		MinimumRequests:  2,
	}

	cb := New(config, logger)
	ctx := context.Background()

	// Execute successful function
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_Execute_FailureThreshold(t *testing.T) {
	logger, _ := log.NewForTest()
	config := Config{
		FailureThreshold: 3,
		RecoveryTimeout:  time.Second,
		MinimumRequests:  2,
	}

	cb := New(config, logger)
	ctx := context.Background()
	testError := errors.New("test error")

	// Execute failing function until threshold is reached
	for i := 0; i < 3; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return testError
		})
		assert.Error(t, err)
	}

	// Circuit breaker should now be open
	assert.Equal(t, StateOpen, cb.GetState())

	// Next execution should fail fast
	err := cb.Execute(ctx, func(ctx context.Context) error {
		t.Fatal("This function should not be called when circuit is open")
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is OPEN")
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	logger, _ := log.NewForTest()
	config := Config{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		MinimumRequests:  2,
	}

	cb := New(config, logger)
	ctx := context.Background()
	testError := errors.New("test error")

	// Force circuit breaker to open
	for range 2 {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return testError
		})
		assert.Error(t, err)
	}

	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Execute successful function - should transition to half-open
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, cb.GetState())

	// Execute another successful function - should close circuit
	err = cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	logger, _ := log.NewForTest()
	config := Config{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		MinimumRequests:  2,
	}

	cb := New(config, logger)
	ctx := context.Background()
	testError := errors.New("test error")

	// Force circuit breaker to open
	for range 2 {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return testError
		})
		assert.Error(t, err)
	}

	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Execute failing function in half-open state - should return to open
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return testError
	})

	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	logger, _ := log.NewForTest()
	config := DefaultConfig()

	cb := New(config, logger)
	ctx := context.Background()

	// Execute a few operations
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err)
	err = cb.Execute(ctx, func(ctx context.Context) error {
		return errors.New("test error")
	})
	assert.Error(t, err)

	stats := cb.GetStats()

	assert.NotNil(t, stats)
	assert.Contains(t, stats, "state")
	assert.Contains(t, stats, "failure_count")
	assert.Contains(t, stats, "request_count")
	assert.Contains(t, stats, "last_fail_time")
}

func TestCircuitBreakerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid failure threshold",
			config: Config{
				FailureThreshold: 0,
				RecoveryTimeout:  time.Second,
				MinimumRequests:  1,
			},
			wantErr: true,
		},
		{
			name: "invalid recovery timeout",
			config: Config{
				FailureThreshold: 5,
				RecoveryTimeout:  500 * time.Millisecond,
				MinimumRequests:  1,
			},
			wantErr: true,
		},
		{
			name: "invalid minimum requests",
			config: Config{
				FailureThreshold: 5,
				RecoveryTimeout:  time.Second,
				MinimumRequests:  0,
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

func TestState_String(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "CLOSED"},
		{StateOpen, "OPEN"},
		{StateHalfOpen, "HALF_OPEN"},
		{State(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	logger, _ := log.NewForTest()
	config := DefaultConfig()

	cb := New(config, logger)
	ctx := context.Background()

	// Test concurrent access to circuit breaker
	done := make(chan bool, 10)

	for range 10 {
		go func() {
			defer func() { done <- true }()

			for j := range 100 {
				err := cb.Execute(ctx, func(ctx context.Context) error {
					if j%10 == 0 {
						return errors.New("test error")
					}
					return nil
				})
				if j%10 == 0 {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	// Circuit breaker should still be functional
	stats := cb.GetStats()
	assert.NotNil(t, stats)
}
