package plugins

import (
	"github.com/evilsocket/islazy/log"
)

// LogPackage contains logs methods
type LogPackage struct{}

// GetLog returns the LogPackage
func GetLog() *LogPackage {
	return &LogPackage{}
}

// Info prints an info line on the logs
func (l *LogPackage) Info(format string, a ...interface{}) {
	log.Info(format, a...)
}

// Error prints an error line on the logs
func (l *LogPackage) Error(format string, a ...interface{}) {
	log.Error(format, a...)
}

// Debug prints a debug line on the logs
func (l *LogPackage) Debug(format string, a ...interface{}) {
	log.Debug(format, a...)
}
