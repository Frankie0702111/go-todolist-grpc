package queue

import (
	"context"
	"fmt"
	"go-todolist-grpc/internal/pkg/log"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (logger *Logger) Print(level int, args ...interface{}) {
	msg := fmt.Sprint(args...)

	switch level {
	case log.LevelDebug:
		log.Debug.Print(msg)
	case log.LevelInfo:
		log.Info.Print(msg)
	case log.LevelWarning:
		log.Warning.Print(msg)
	case log.LevelError:
		log.Error.Print(msg)
	default:
		log.Info.Print(msg)
	}
}

func (logger *Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Debug.Print(msg)
}

func (logger *Logger) Debug(args ...interface{}) {
	logger.Print(log.LevelDebug, args...)
}

func (logger *Logger) Info(args ...interface{}) {
	logger.Print(log.LevelInfo, args...)
}

func (logger *Logger) Warn(args ...interface{}) {
	logger.Print(log.LevelWarning, args...)
}

func (logger *Logger) Error(args ...interface{}) {
	logger.Print(log.LevelError, args...)
}

func (logger *Logger) Fatal(args ...interface{}) {
	logger.Print(log.LevelError, args...)
}
