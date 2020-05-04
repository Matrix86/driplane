package plugins

import (
	"github.com/evilsocket/islazy/log"
)

type LogPackage struct{}

func GetLog() *LogPackage {
	return &LogPackage{}
}

func (l *LogPackage) Info(format string, a ...interface{}) {
	log.Info(format, a...)
}

func (l *LogPackage) Error(format string, a ...interface{}) {
	log.Error(format, a...)
}

func (l *LogPackage) Debug(format string, a ...interface{}) {
	log.Debug(format, a...)
}
