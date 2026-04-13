package util

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
}

const (
	LevelInfo = iota
	LevelWarn
	LevelError
	LevelFatal
)

// Logger represent logger
type Logger struct {
	logger *logrus.Entry
}

func NewLogger(component string) *Logger {
	return &Logger{
		logger: logrus.WithField("component", component),
	}
}

// Info [#TODO](should add some comments)
func (l *Logger) Info(args ...interface{}) {
	l.logger.Info(args...)
}
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *Logger) Warn(args ...any) {
	l.logger.Warn(args...)
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.logger.Error(args...)
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}
