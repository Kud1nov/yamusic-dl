// Package logger provides logging functionality using zerolog.
package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger wrapper around zerolog.Logger
type Logger struct {
	logger zerolog.Logger
	level  zerolog.Level
}

// New creates a new logger instance.
// If verbose=true, debug level logging will be enabled.
func New(verbose bool) *Logger {
	// Configure logging level
	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	}

	// Configure output
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}

	// Create and configure logger
	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &Logger{
		logger: logger,
		level:  level,
	}
}

// Debug logs debug messages
func (l *Logger) Debug(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

// Info logs informational messages
func (l *Logger) Info(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}

// IsDebug returns true if debug logging is enabled
func (l *Logger) IsDebug() bool {
	return l.level == zerolog.DebugLevel
}

// Logger returns the underlying zerolog.Logger for more flexible usage
func (l *Logger) Logger() zerolog.Logger {
	return l.logger
}
