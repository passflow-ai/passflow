package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the logging interface.
type Logger interface {
	Initialize() error
	Info(message, id string, args ...interface{})
	Warn(message, id string, args ...interface{})
	Debug(message, id string, args ...interface{})
	Error(message, id string, args ...interface{})
}

type logger struct {
	writeFile   bool
	loggers     map[zapcore.Level]*zap.SugaredLogger
	projectPath string
}

// NewLogger creates a new Logger instance.
func NewLogger(projectPath string, writeFile bool) Logger {
	return &logger{
		writeFile:   writeFile,
		loggers:     make(map[zapcore.Level]*zap.SugaredLogger),
		projectPath: projectPath,
	}
}

// Initialize sets up the logger with the specified configuration.
func (l *logger) Initialize() error {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "severity"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	levels := []zapcore.Level{
		zap.DebugLevel,
		zap.InfoLevel,
		zap.WarnLevel,
		zap.ErrorLevel,
	}

	for _, level := range levels {
		outputs := []string{"stdout"}

		if l.writeFile {
			logFilePath, err := l.getLogFilePath(level.String())
			if err != nil {
				return fmt.Errorf("failed to create log file path: %w", err)
			}
			outputs = append(outputs, logFilePath)
		}

		cfg := zap.Config{
			Level:             zap.NewAtomicLevelAt(level),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: true,
			Encoding:          "json",
			EncoderConfig:     encoderConfig,
			OutputPaths:       outputs,
			ErrorOutputPaths:  []string{"stderr"},
		}

		zLogger, err := cfg.Build()
		if err != nil {
			return fmt.Errorf("failed to build logger for level %s: %w", level.String(), err)
		}

		l.loggers[level] = zLogger.Sugar()
	}

	return nil
}

// Info logs an informational message.
func (l *logger) Info(message, id string, args ...interface{}) {
	l.log(zapcore.InfoLevel, message, id, args...)
}

// Warn logs a warning message.
func (l *logger) Warn(message, id string, args ...interface{}) {
	l.log(zapcore.WarnLevel, message, id, args...)
}

// Debug logs a debug message.
func (l *logger) Debug(message, id string, args ...interface{}) {
	l.log(zapcore.DebugLevel, message, id, args...)
}

// Error logs an error message.
func (l *logger) Error(message, id string, args ...interface{}) {
	l.log(zapcore.ErrorLevel, message, id, args...)
}

func (l *logger) log(level zapcore.Level, message, id string, args ...interface{}) {
	if id == "" {
		id = uuid.New().String()
	}

	logger := l.loggers[level]
	if logger == nil {
		fmt.Printf("{\"severity\":\"%s\",\"message\":\"%s\",\"id\":\"%s\"}\n",
			level.String(), message, id)
		return
	}

	formattedMessage := message
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(strings.ToLower(message), args...)
	}

	switch level {
	case zapcore.DebugLevel:
		logger.Debugw(formattedMessage, "trace_id", id)
	case zapcore.InfoLevel:
		logger.Infow(formattedMessage, "trace_id", id)
	case zapcore.WarnLevel:
		logger.Warnw(formattedMessage, "trace_id", id)
	case zapcore.ErrorLevel:
		logger.Errorw(formattedMessage, "trace_id", id)
	}
}

func (l *logger) getLogFilePath(level string) (string, error) {
	if l.projectPath == "" {
		l.projectPath = "./logs"
	}

	if err := os.MkdirAll(l.projectPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create logs directory: %w", err)
	}

	return filepath.Join(l.projectPath, fmt.Sprintf("%s.log", level)), nil
}
