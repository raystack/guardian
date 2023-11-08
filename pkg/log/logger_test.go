package log

import (
	"context"
	"testing"

	"github.com/goto/guardian/pkg/log/mocks"
)

func TestLogger(t *testing.T) {
	saltLogger := new(mocks.SaltLogger)
	l := NewCtxLoggerWithSaltLogger(saltLogger, []string{"ctx-key-1"})

	t.Run("empty context", func(t *testing.T) {
		t.Run("Debug", func(t *testing.T) {
			saltLogger.EXPECT().Debug("this is a test debug message", []interface{}{"keys", "test-value"}...).Once()
			l.Debug(nil, "this is a test debug message", "keys", "test-value")
			saltLogger.AssertExpectations(t)
		})

		t.Run("Info", func(t *testing.T) {
			saltLogger.EXPECT().Info("this is a test info message", []interface{}{"keys", "test-value"}...).Once()
			l.Info(nil, "this is a test info message", "keys", "test-value")
			saltLogger.AssertExpectations(t)
		})

		t.Run("Warn", func(t *testing.T) {
			saltLogger.EXPECT().Warn("this is a test warn message", []interface{}{"keys", "test-value"}...).Once()
			l.Warn(nil, "this is a test warn message", "keys", "test-value")
			saltLogger.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			saltLogger.EXPECT().Error("this is a test error message", []interface{}{"keys", "test-value"}...).Once()
			l.Error(nil, "this is a test error message", "keys", "test-value")
			saltLogger.AssertExpectations(t)
		})

		t.Run("Fatal", func(t *testing.T) {
			saltLogger.EXPECT().Fatal("this is a test fatal message", []interface{}{"keys", "test-value"}...).Once()
			l.Fatal(nil, "this is a test fatal message", "keys", "test-value")
			saltLogger.AssertExpectations(t)
		})
	})

	t.Run("context with keys", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "ctx-key-1", "ctx-value-1")
		t.Run("Debug", func(t *testing.T) {
			saltLogger.EXPECT().Debug("this is a test debug message", []interface{}{"key1", "test-value1", "ctx-key-1", "ctx-value-1"}...).Once()
			l.Debug(ctx, "this is a test debug message", "key1", "test-value1")
			saltLogger.AssertExpectations(t)
		})

		t.Run("Info", func(t *testing.T) {
			saltLogger.EXPECT().Info("this is a test info message", []interface{}{"key1", "test-value1", "ctx-key-1", "ctx-value-1"}...).Once()
			l.Info(ctx, "this is a test info message", "key1", "test-value1")
			saltLogger.AssertExpectations(t)
		})

		t.Run("Warn", func(t *testing.T) {
			saltLogger.EXPECT().Warn("this is a test warn message", []interface{}{"key1", "test-value1", "ctx-key-1", "ctx-value-1"}...).Once()
			l.Warn(ctx, "this is a test warn message", "key1", "test-value1")
			saltLogger.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			saltLogger.EXPECT().Error("this is a test error message", []interface{}{"key1", "test-value1", "ctx-key-1", "ctx-value-1"}...).Once()
			l.Error(ctx, "this is a test error message", "key1", "test-value1")
			saltLogger.AssertExpectations(t)
		})

		t.Run("Fatal", func(t *testing.T) {
			saltLogger.EXPECT().Fatal("this is a test fatal message", []interface{}{"key1", "test-value1", "ctx-key-1", "ctx-value-1"}...).Once()
			l.Fatal(ctx, "this is a test fatal message", "key1", "test-value1")
			saltLogger.AssertExpectations(t)
		})
	})
}
