package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	l *slog.Logger
}

func NewLogger() *Logger {
	return &Logger{
		l: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

func (l *Logger) Error(msg string, args ...any) {
	l.l.Error(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.l.Warn(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.l.Info(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.l.Debug(msg, args...)
}
