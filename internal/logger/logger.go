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

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...interface{}) {
	l.logger.Warn().Msgf(format, v...)
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.logger.Fatal().Msgf(format, v...)
}

// Infof logs formatted informational messages
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

// Errorf logs formatted error messages
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}

// Debugf logs formatted debug messages
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

// Warnf logs formatted warning messages
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logger.Warn().Msgf(format, v...)
}

// Fatalf logs formatted fatal messages and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatal().Msgf(format, v...)
}
