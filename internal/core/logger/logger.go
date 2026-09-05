package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Logger struct {
	*slog.Logger
	file *os.File
}

type contextKey string

const loggerContextKey contextKey = "log"

func FromContext(ctx context.Context) *Logger {
	log, ok := TryFromContext(ctx)
	if !ok {
		panic("no logger in context")
	}

	return log
}

func FromContextOr(ctx context.Context, fallback *Logger) *Logger {
	if log, ok := TryFromContext(ctx); ok {
		return log
	}

	return fallback
}

func TryFromContext(ctx context.Context) (*Logger, bool) {
	log, ok := ctx.Value(loggerContextKey).(*Logger)

	return log, ok
}

func WithContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}

func NewLogger(config Config) (*Logger, error) {
	lvl, err := parseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(config.Folder, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir log folder: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
	logFilePath := filepath.Join(config.Folder, fmt.Sprintf("%s.log", timestamp))

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	out := io.MultiWriter(os.Stdout, logFile)

	handler := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: true,
	})

	return &Logger{
		Logger: slog.New(handler),
		file:   logFile,
	}, nil
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		file:   l.file,
	}
}

func (l *Logger) Close() {
	if err := l.file.Close(); err != nil {
		fmt.Println("failed to close application logger:", err)
	}
}

func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unmarshal log level: %q", s)
	}
}
