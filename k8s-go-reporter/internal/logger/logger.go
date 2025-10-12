package logger

import "log"

// Logger defines the interface for application logging
// This follows the Dependency Inversion Principle - components depend on this interface, not concrete implementations
type Logger interface {
	// Info logs an informational message
	Info(msg string)
	// Infof logs a formatted informational message
	Infof(format string, args ...interface{})
	// Error logs an error message
	Error(msg string, err error)
	// Errorf logs a formatted error message
	Errorf(format string, args ...interface{})
}

// StdLogger implements Logger using Go's standard library log package
type StdLogger struct{}

// NewStdLogger creates a new standard library logger
func NewStdLogger() Logger {
	return &StdLogger{}
}

// Info logs an informational message
func (l *StdLogger) Info(msg string) {
	log.Println(msg)
}

// Infof logs a formatted informational message
func (l *StdLogger) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// Error logs an error message with the associated error
func (l *StdLogger) Error(msg string, err error) {
	log.Printf("ERROR: %s: %v", msg, err)
}

// Errorf logs a formatted error message
func (l *StdLogger) Errorf(format string, args ...interface{}) {
	log.Printf("ERROR: "+format, args...)
}
