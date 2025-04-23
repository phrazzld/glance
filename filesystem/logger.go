// Package filesystem provides functionality for scanning, reading, and managing
// filesystem operations in the glance application.
package filesystem

import (
	"github.com/sirupsen/logrus"
)

// log is the package-level logger that will be used throughout the filesystem package.
// It is initialized with the standard logrus logger by default, but can be replaced with
// SetLogger.
var log logrus.FieldLogger = logrus.StandardLogger()

// SetLogger sets the package-level logger used throughout the filesystem package.
// This function allows injecting a custom logger from outside the package.
//
// Parameters:
//   - logger: The FieldLogger instance to use for all logging in the filesystem package
func SetLogger(logger logrus.FieldLogger) {
	if logger != nil {
		log = logger
	}
}

// IsLevelEnabled checks if the current logger has the specified level enabled.
// This provides a consistent way to check log levels throughout the package.
//
// Parameters:
//   - level: The log level to check
//
// Returns:
//   - true if the level is enabled, false otherwise
func IsLevelEnabled(level logrus.Level) bool {
	// If the logger is a *logrus.Logger, we can directly call its IsLevelEnabled method
	if l, ok := log.(*logrus.Logger); ok {
		return l.IsLevelEnabled(level)
	}

	// Otherwise, if it's a custom logger, we have no direct way of checking.
	// In this case, we'll assume the level is enabled, and the actual logger's
	// filtering will handle it.
	return true
}
