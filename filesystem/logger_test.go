package filesystem

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestSetLogger verifies that the SetLogger function correctly changes the package logger
func TestSetLogger(t *testing.T) {
	// Backup the original logger to restore after test
	originalLogger := log
	defer func() {
		log = originalLogger
	}()

	// Create a test logger
	testLogger := logrus.New()
	var buf bytes.Buffer
	testLogger.SetOutput(&buf)
	testLogger.SetLevel(logrus.DebugLevel)

	// Set the test logger
	SetLogger(testLogger)

	// Verify log is set to the test logger
	assert.Equal(t, testLogger, log)

	// Use log to check output
	log.Debug("test debug message")
	assert.Contains(t, buf.String(), "test debug message")

	// Test nil logger doesn't panic
	SetLogger(nil)
	// log should remain unchanged if nil is passed
	assert.Equal(t, testLogger, log)
}

// TestIsLevelEnabled verifies the IsLevelEnabled function
func TestIsLevelEnabled(t *testing.T) {
	// Backup the original logger to restore after test
	originalLogger := log
	defer func() {
		log = originalLogger
	}()

	// Test with standard logrus logger
	testLogger := logrus.New()
	testLogger.SetLevel(logrus.InfoLevel)
	SetLogger(testLogger)

	// Debug level should be disabled
	assert.False(t, IsLevelEnabled(logrus.DebugLevel))
	// Info level should be enabled
	assert.True(t, IsLevelEnabled(logrus.InfoLevel))
	// Warn level should be enabled
	assert.True(t, IsLevelEnabled(logrus.WarnLevel))
	// Error level should be enabled
	assert.True(t, IsLevelEnabled(logrus.ErrorLevel))

	// Change log level to debug
	testLogger.SetLevel(logrus.DebugLevel)
	// Now debug level should be enabled
	assert.True(t, IsLevelEnabled(logrus.DebugLevel))

	// Test with custom struct implementing FieldLogger
	// This should always return true since we can't directly check log level
	type customLogger struct {
		logrus.FieldLogger
	}
	SetLogger(&customLogger{FieldLogger: testLogger})
	assert.True(t, IsLevelEnabled(logrus.DebugLevel))
	assert.True(t, IsLevelEnabled(logrus.InfoLevel))
	assert.True(t, IsLevelEnabled(logrus.WarnLevel))
	assert.True(t, IsLevelEnabled(logrus.ErrorLevel))
}
