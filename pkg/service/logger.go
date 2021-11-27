package service

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	WithContext(ctx context.Context) Logger
	WithError(err error) Logger
	WithFields(map[string]interface{}) Logger
}

type loggerWrapper struct {
	l *logrus.Entry
}

func newLoggerWrapper(name string) *loggerWrapper {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetFormatter(&ecslogrus.Formatter{})

	return &loggerWrapper{
		l: l.WithField("service", name),
	}
}

func (lw *loggerWrapper) Debugf(format string, args ...interface{}) {
	lw.l.Debugf(format, args)
}

func (lw *loggerWrapper) Infof(format string, args ...interface{}) {
	lw.l.Infof(format, args)
}

func (lw *loggerWrapper) Warnf(format string, args ...interface{}) {
	lw.l.Warnf(format, args)
}

func (lw *loggerWrapper) Errorf(format string, args ...interface{}) {
	lw.l.Errorf(format, args)
}

func (lw *loggerWrapper) Debug(args ...interface{}) {
	lw.l.Debug(args)
}

func (lw *loggerWrapper) Info(args ...interface{}) {
	lw.l.Info(args)
}

func (lw *loggerWrapper) Warn(args ...interface{}) {
	lw.l.Warn(args)
}

func (lw *loggerWrapper) Error(args ...interface{}) {
	lw.l.Error(args)
}

func (lw *loggerWrapper) WithContext(ctx context.Context) Logger {
	return &loggerWrapper{
		l: lw.l.WithContext(ctx),
	}
}

func (lw *loggerWrapper) WithError(err error) Logger {
	return &loggerWrapper{
		l: lw.l.WithError(err),
	}
}
func (lw *loggerWrapper) WithFields(fields map[string]interface{}) Logger {
	return &loggerWrapper{
		l: lw.l.WithFields(fields),
	}
}
