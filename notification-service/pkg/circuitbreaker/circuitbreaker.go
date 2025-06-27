package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed indicates that the circuit breaker is closed (normal operation)
	StateClosed State = iota
	// StateOpen indicates that the circuit breaker is open (failing fast)
	StateOpen
	// StateHalfOpen indicates that the circuit breaker is half-open (testing)
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config defines circuit breaker configuration
type Config struct {
	FailureThreshold int           `yaml:"failure_threshold" json:"failureThreshold"`
	RecoveryTimeout  time.Duration `yaml:"recovery_timeout" json:"recoveryTimeout"`
	MinimumRequests  int           `yaml:"minimum_requests" json:"minimumRequests"`
}

// DefaultConfig returns a default circuit breaker configuration
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		RecoveryTimeout:  60 * time.Second,
		MinimumRequests:  3,
	}
}

// Validate validates the circuit breaker configuration
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.FailureThreshold, validation.Required, validation.Min(1)),
		validation.Field(&c.RecoveryTimeout, validation.Required, validation.Min(time.Second)),
		validation.Field(&c.MinimumRequests, validation.Required, validation.Min(1)),
	)
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config       Config
	state        State
	failureCount int
	requestCount int
	lastFailTime time.Time
	mutex        sync.RWMutex
	logger       log.Logger
}

// New creates a new circuit breaker
func New(config Config, logger log.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		logger: logger,
	}
}

// Execute executes a function through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check if we can execute the function
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is OPEN, failing fast")
	}

	// Execute the function
	err := fn(ctx)

	// Record the result
	cb.recordResult(err)

	return err
}

// canExecute determines if the function can be executed
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if recovery timeout has passed
		if time.Since(cb.lastFailTime) >= cb.config.RecoveryTimeout {
			cb.state = StateHalfOpen
			cb.requestCount = 0
			cb.logger.Info("Circuit breaker transitioning to HALF_OPEN state")
			return true
		}
		return false
	case StateHalfOpen:
		return cb.requestCount < cb.config.MinimumRequests
	default:
		return false
	}
}

// recordResult records the result of function execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.requestCount++

	if err != nil {
		cb.failureCount++
		cb.lastFailTime = time.Now()

		switch cb.state {
		case StateClosed:
			if cb.failureCount >= cb.config.FailureThreshold {
				cb.state = StateOpen
				cb.logger.With(context.Background(), "failure_count", cb.failureCount).Error("Circuit breaker opening due to failure threshold")
			}
		case StateHalfOpen:
			cb.state = StateOpen
			cb.logger.Error("Circuit breaker returning to OPEN state after failure in HALF_OPEN")
		}
	} else {
		// Success
		switch cb.state {
		case StateHalfOpen:
			if cb.requestCount >= cb.config.MinimumRequests {
				cb.state = StateClosed
				cb.failureCount = 0
				cb.requestCount = 0
				cb.logger.Info("Circuit breaker closing after successful requests in HALF_OPEN")
			}
		case StateClosed:
			// Reset failure count on successful request
			if cb.failureCount > 0 {
				cb.failureCount = 0
			}
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"state":          cb.state.String(),
		"failure_count":  cb.failureCount,
		"request_count":  cb.requestCount,
		"last_fail_time": cb.lastFailTime,
	}
}
