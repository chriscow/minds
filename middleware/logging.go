package middleware

import (
	"log/slog"
	"time"

	"github.com/chriscow/minds"
)

// LoggingOptions defines configuration for logging middleware.
type LoggingOptions struct {
	Logger      *slog.Logger
	LogMessages bool
	LogMetadata bool
	LogLevels   LogLevels
}

// LogLevels specifies log levels for different events.
type LogLevels struct {
	Entry slog.Level
	Exit  slog.Level
	Error slog.Level
}

// NewLoggingOptions creates default logging configuration.
func NewLoggingOptions() *LoggingOptions {
	return &LoggingOptions{
		Logger:      slog.Default(),
		LogMessages: true,
		LogMetadata: true,
		LogLevels: LogLevels{
			Entry: slog.LevelInfo,
			Exit:  slog.LevelInfo,
			Error: slog.LevelError,
		},
	}
}

// LoggingOption defines a configuration function for logging middleware.
type LoggingOption func(*LoggingOptions)

// WithLogger sets a custom logger.
func WithLogger(logger *slog.Logger) LoggingOption {
	return func(o *LoggingOptions) {
		o.Logger = logger
	}
}

// WithLogMessages configures message logging.
func WithLogMessages(enabled bool) LoggingOption {
	return func(o *LoggingOptions) {
		o.LogMessages = enabled
	}
}

// WithLogMetadata configures metadata logging.
func WithLogMetadata(enabled bool) LoggingOption {
	return func(o *LoggingOptions) {
		o.LogMetadata = enabled
	}
}

// WithLogLevels sets custom log levels for different events.
func WithLogLevels(entry, exit, errorLevel slog.Level) LoggingOption {
	return func(o *LoggingOptions) {
		o.LogLevels.Entry = entry
		o.LogLevels.Exit = exit
		o.LogLevels.Error = errorLevel
	}
}

// Logging creates a middleware that logs thread execution details.
//
// The middleware provides configurable logging with options to:
//   - Use a custom logger
//   - Enable/disable message and metadata logging
//   - Set custom log levels for different events
//
// Example usage:
//
//	flow.Use(Logging("api_handler",
//	  WithLogger(customLogger),
//	  WithLogMessages(false),
//	  WithLogLevels(slog.LevelDebug, slog.LevelInfo, slog.LevelError)
//	))
func Logging(name string, opts ...LoggingOption) minds.Middleware {
	// Apply configuration options
	options := NewLoggingOptions()
	for _, opt := range opts {
		opt(options)
	}

	return minds.MiddlewareFunc(func(next minds.ThreadHandler) minds.ThreadHandler {
		return minds.ThreadHandlerFunc(func(tc minds.ThreadContext, _ minds.ThreadHandler) (minds.ThreadContext, error) {
			// Prepare base logging attributes
			baseAttrs := []any{
				"handler", name,
				"thread_id", tc.UUID(),
			}

			// Log entry
			logEntry(options, tc, baseAttrs, "entering handler", options.LogLevels.Entry)

			// Execute handler with timing
			start := time.Now()
			result, err := next.HandleThread(tc, nil)
			duration := time.Since(start)

			// Prepare result attributes
			resultAttrs := prepareResultAttributes(baseAttrs, result, options, duration)

			// Handle and log errors
			if err != nil {
				resultAttrs = append(resultAttrs, "error", err.Error())
				logEntry(options, tc, resultAttrs, "handler error", options.LogLevels.Error)
				return result, err
			}

			// Log successful completion
			logEntry(options, tc, resultAttrs, "exiting handler", options.LogLevels.Exit)

			return result, nil
		})
	})
}

// logEntry handles logging with configurable options
func logEntry(options *LoggingOptions, tc minds.ThreadContext, attrs []any, msg string, level slog.Level) {
	options.Logger.LogAttrs(tc.Context(), level, msg, slog.Group("thread", attrs...))
}

// prepareResultAttributes builds logging attributes for the handler result
func prepareResultAttributes(baseAttrs []any, result minds.ThreadContext, options *LoggingOptions, duration time.Duration) []any {
	attrs := make([]any, len(baseAttrs))
	copy(attrs, baseAttrs)

	attrs = append(attrs, "duration", duration)

	if options.LogMessages {
		attrs = append(attrs, "messages", result.Messages())
	}
	if options.LogMetadata {
		attrs = append(attrs, "metadata", result.Metadata())
	}

	return attrs
}
