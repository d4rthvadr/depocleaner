package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

type LogLevel string

const (
	LevelDebug      LogLevel = "debug"
	LevelInfo       LogLevel = "info"
	LevelWarning    LogLevel = "warning"
	LevelError      LogLevel = "error"
	DefaultLogLevel LogLevel = "info"
)

func New(logPath string, level LogLevel) (*Logger, error) {

	// Ensure log directory exists
	logDir := filepath.Dir(logPath)
	if err := os.Mkdir(logDir, 0755); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Parse log level
	logLevel, err := zapcore.ParseLevel(string(level))
	if err != nil {
		logLevel = zapcore.InfoLevel
	}

	// Configure zap logger
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		logPath,
		"stdout", // Also write logs to stdout
	}

	cfg.ErrorOutputPaths = []string{
		logPath,
		"stderr",
	}

	cfg.Encoding = "json"
	cfg.Level = zap.NewAtomicLevelAt(logLevel)

	// Customize time encoding for consistency and reading ease
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("building logger: %w", err)
	}

	return &Logger{logger.Sugar()}, nil
}

// NewConsoleOnly creates a logger that outputs only to the console (stdout)
// at the development log level.
func NewConsoleOnly() *Logger {
	logger, _ := zap.NewDevelopment()
	return &Logger{logger.Sugar()}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var zapFields []interface{}
	for k, v := range fields {
		zapFields = append(zapFields, k, v)
	}
	return &Logger{l.SugaredLogger.With(zapFields...)}
}

// Close and flushes any buffered log entries
func (l *Logger) Close() error {
	return l.Sync()
}
