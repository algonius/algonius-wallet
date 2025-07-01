// Package logger provides a mock logger for unit testing.
package logger

import (
	"sync"

	"go.uber.org/zap"
)

// MockLogger implements Logger for testing.
type MockLogger struct {
	mu     sync.Mutex
	Infos  []string
	Debugs []string
	Errors []string
}

func (l *MockLogger) Named(name string) Logger {
	return l
}

func (l *MockLogger) Sync() error {
	return nil
}

func (l *MockLogger) Info(msg string, fields ...zap.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Infos = append(l.Infos, msg)
}
func (l *MockLogger) Debug(msg string, fields ...zap.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Debugs = append(l.Debugs, msg)
}
func (l *MockLogger) Error(msg string, fields ...zap.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Errors = append(l.Errors, msg)
}

func (l *MockLogger) Warn(msg string, fields ...zap.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Errors = append(l.Errors, msg)
}

func (l *MockLogger) With(fields ...zap.Field) Logger {
	return l
}

// NewMockLogger creates a new MockLogger.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}
