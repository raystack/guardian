package log

import (
	"context"
	"io"

	saltLog "github.com/goto/salt/log"
)

type Logger interface {

	// Debug level message with alternating key/value pairs
	// key should be string, value could be anything printable
	Debug(ctx context.Context, msg string, args ...interface{})

	// Info level message with alternating key/value pairs
	// key should be string, value could be anything printable
	Info(ctx context.Context, msg string, args ...interface{})

	// Warn level message with alternating key/value pairs
	// key should be string, value could be anything printable
	Warn(ctx context.Context, msg string, args ...interface{})

	// Error level message with alternating key/value pairs
	// key should be string, value could be anything printable
	Error(ctx context.Context, msg string, args ...interface{})

	// Fatal level message with alternating key/value pairs
	// key should be string, value could be anything printable
	Fatal(ctx context.Context, msg string, args ...interface{})

	// Level returns priority level for which this logger will filter logs
	Level() string

	// Writer used to print logs
	Writer() io.Writer
}

type CtxLogger struct {
	log  saltLog.Logger
	keys []string
}

// NewCtxLoggerWithSaltLogger returns a logger that will add context value to the log message, wrapped with saltLog.Logger
func NewCtxLoggerWithSaltLogger(log saltLog.Logger, ctxKeys []string) *CtxLogger {
	return &CtxLogger{log: log, keys: ctxKeys}
}

// NewCtxLogger returns a logger that will add context value to the log message
func NewCtxLogger(logLevel string, ctxKeys []string) *CtxLogger {
	saltLogger := saltLog.NewLogrus(saltLog.LogrusWithLevel(logLevel))
	return NewCtxLoggerWithSaltLogger(saltLogger, ctxKeys)
}

func (l *CtxLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.log.Debug(msg, l.addCtxToArgs(ctx, args)...)
}

func (l *CtxLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.log.Info(msg, l.addCtxToArgs(ctx, args)...)
}

func (l *CtxLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.log.Warn(msg, l.addCtxToArgs(ctx, args)...)
}

func (l *CtxLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.log.Error(msg, l.addCtxToArgs(ctx, args)...)
}

func (l *CtxLogger) Fatal(ctx context.Context, msg string, args ...interface{}) {
	l.log.Fatal(msg, l.addCtxToArgs(ctx, args)...)
}

func (l *CtxLogger) Level() string {
	return l.log.Level()
}

func (l *CtxLogger) Writer() io.Writer {
	return l.log.Writer()
}

// addCtxToArgs adds context value to the existing args slice as key/value pair
func (l *CtxLogger) addCtxToArgs(ctx context.Context, args []interface{}) []interface{} {
	if ctx == nil {
		return args
	}

	for _, key := range l.keys {
		if val, ok := ctx.Value(key).(string); ok {
			args = append(args, key, val)
		}
	}

	return args
}
