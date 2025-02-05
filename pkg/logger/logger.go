package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

var Log *zap.Logger

func NewLogger(logPath string) (*zap.Logger, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{logPath, "stdout"},
		ErrorOutputPaths: []string{logPath, "stderr"},
	}

	logger, err := config.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}

	Log = logger
	return logger, nil
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}
