package retry

import (
	"time"
)

type Option func(*Options)

type BackoffStrategy func(attempt int) time.Duration
type Criteria func(err error) bool

type Options struct {
	Attempts         int
	Backoff          BackoffStrategy
	ShouldRetry      Criteria
	PropagateTimeout bool
}

func NewDefaultOptions() *Options {
	return &Options{
		Attempts:         3,
		Backoff:          DefaultBackoff(0), // No delay by default
		ShouldRetry:      DefaultCriteria,
		PropagateTimeout: true,
	}
}

// DefaultBackoff provides a simple constant delay
func DefaultBackoff(delay time.Duration) BackoffStrategy {
	return func(_ int) time.Duration {
		return delay
	}
}

// DefaultCriteria retries on all errors
func DefaultCriteria(err error) bool {
	return err != nil
}

func WithAttempts(attempts int) Option {
	return func(config *Options) {
		config.Attempts = attempts
	}
}

func WithBackoff(strategy BackoffStrategy) Option {
	return func(config *Options) {
		config.Backoff = strategy
	}
}

func WithRetryCriteria(criteria Criteria) Option {
	return func(config *Options) {
		config.ShouldRetry = criteria
	}
}

func WithoutTimeoutPropagation() Option {
	return func(config *Options) {
		config.PropagateTimeout = false
	}
}
