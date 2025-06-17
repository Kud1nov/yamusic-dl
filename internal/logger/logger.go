// Package logger предоставляет функциональность для логирования с использованием zerolog.
package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger обертка вокруг zerolog.Logger
type Logger struct {
	logger zerolog.Logger
}

// New создает новый экземпляр логгера.
// Если verbose=true, будет включен отладочный уровень логирования.
func New(verbose bool) *Logger {
	// Настраиваем уровень логирования
	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	}

	// Настраиваем вывод
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}

	// Создаем и настраиваем логгер
	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &Logger{
		logger: logger,
	}
}

// Debug логирует отладочные сообщения
func (l *Logger) Debug(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

// Info логирует информационные сообщения
func (l *Logger) Info(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

// Error логирует сообщения об ошибках
func (l *Logger) Error(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}

// Logger возвращает базовый zerolog.Logger для более гибкого использования
func (l *Logger) Logger() zerolog.Logger {
	return l.logger
}
