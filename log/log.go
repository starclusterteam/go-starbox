package log

import (
	"github.com/sirupsen/logrus"

	"github.com/starcluster/go-starbox/config"
)

var baseLogger = logger{logrus.NewEntry(logrus.StandardLogger())}

// Logger returns underlying logger.
func Logger() Interface {
	return baseLogger
}

// LogrusLogger returns the logrus logger used by the underlying logger.
func LogrusLogger() *logrus.Logger {
	return baseLogger.Entry.Logger
}

func init() {
	level, err := logrus.ParseLevel(config.String("LOG_SEVERITY", "info"))
	if err != nil {
		logrus.Fatalf("Failed to parse level from LOG_SEVERITY environment variable: %v", err)
	}

	logrus.SetLevel(level)
	if !config.FetchGoEnv().Development {
		// Use structured formatter.
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

// Interface is logger interface.
type Interface interface {
	logrus.FieldLogger
	With(key string, value interface{}) Interface
}

type logger struct {
	*logrus.Entry
}

func (l logger) With(key string, value interface{}) Interface {
	return logger{l.Entry.WithField(key, value)}
}

// Debugf logs to the DEBUG log.
func Debugf(format string, args ...interface{}) {
	baseLogger.Debugf(format, args...)
}

// Infof logs to the INFO log.
func Infof(format string, args ...interface{}) {
	baseLogger.Infof(format, args...)
}

// Warningf logs to the WARN and INFO logs.
func Warningf(format string, args ...interface{}) {
	baseLogger.Warnf(format, args...)
}

// Errorf logs to the ERROR, WARN, and INFO logs.
func Errorf(format string, args ...interface{}) {
	baseLogger.Errorf(format, args...)
}

// Fatalf logs an error message and then exits
func Fatalf(format string, args ...interface{}) {
	baseLogger.Fatalf(format, args...)
}

// PanicExit logs error and exits
func PanicExit(err error) {
	if err != nil {
		Fatalf("%+v", err)
	}
}
