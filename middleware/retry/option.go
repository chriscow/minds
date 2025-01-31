package retry

import (
	"time"

	"github.com/chriscow/minds"
)

// Option defines a retry configuration function.
type Option func(*Options)

// BackoffStrategy defines how delay increases between attempts.
type BackoffStrategy func(attempt int) time.Duration

// Criteria defines when a retry should be attempted.
type Criteria func(tc minds.ThreadContext, attempt int, err error) bool

// Options defines retry behavior.
type Options struct {
	Attempts         int
	Backoff          BackoffStrategy
	ShouldRetry      Criteria
	PropagateTimeout bool
}

// NewDefaultOptions returns default retry options.
func NewDefaultOptions() *Options {
	return &Options{
		Attempts:         3,
		Backoff:          DefaultBackoff(0),
		ShouldRetry:      DefaultCriteria,
		PropagateTimeout: true,
	}
}

// DefaultBackoff provides a simple constant delay.
func DefaultBackoff(delay time.Duration) BackoffStrategy {
	return func(_ int) time.Duration {
		return delay
	}
}

// DefaultCriteria retries on all errors.
func DefaultCriteria(tc minds.ThreadContext, attempt int, err error) bool {
	return err != nil
}

// WithAttempts configures max retry attempts.
func WithAttempts(attempts int) Option {
	return func(config *Options) {
		config.Attempts = attempts
	}
}

// WithBackoff sets a custom backoff strategy.
func WithBackoff(strategy BackoffStrategy) Option {
	return func(config *Options) {
		config.Backoff = strategy
	}
}

// WithRetryCriteria sets a custom retry criteria.
func WithRetryCriteria(criteria Criteria) Option {
	return func(config *Options) {
		config.ShouldRetry = criteria
	}
}

// WithoutTimeoutPropagation disables context timeout propagation.
func WithoutTimeoutPropagation() Option {
	return func(config *Options) {
		config.PropagateTimeout = false
	}
}
