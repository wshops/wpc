package wpclogger

import (
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/wshops/zlog"
	"go.uber.org/zap"
)

type BlackHoleLogger struct {
	zapLog *zap.SugaredLogger
	le     *BhLogEntry
}

func NewBlackholeLogger(zapLogger *zap.SugaredLogger) *BlackHoleLogger {
	return &BlackHoleLogger{
		zapLog: zapLogger,
		le:     NewBhLogEntry(zapLogger),
	}
}

func (w *BlackHoleLogger) SubLogger(fields log.Fields) log.Logger {
	w.zapLog.With("fields", fields)
	return w
}

func (w *BlackHoleLogger) WithFields(fields log.Fields) log.Entry {
	return w.le.WithFields(fields)
}

func (w *BlackHoleLogger) WithField(name string, value interface{}) log.Entry {
	return w.le.WithField(name, value)
}

func (w *BlackHoleLogger) WithError(err error) log.Entry {
	return w.le.WithField("error", err)
}

func (w *BlackHoleLogger) Debug(args ...interface{}) {
	return
}

func (w *BlackHoleLogger) Info(args ...interface{}) {
	return
}

func (w *BlackHoleLogger) Warn(args ...interface{}) {
	zlog.Log().Warn(args...)
}

func (w *BlackHoleLogger) Error(args ...interface{}) {
	zlog.Log().Error(args...)
}

func (w *BlackHoleLogger) Debugf(format string, args ...interface{}) {
	return
}

func (w *BlackHoleLogger) Infof(format string, args ...interface{}) {
	return
}

func (w *BlackHoleLogger) Warnf(format string, args ...interface{}) {
	zlog.Log().Warnf(format, args...)
}

func (w *BlackHoleLogger) Errorf(format string, args ...interface{}) {
	zlog.Log().Errorf(format, args...)
}

type BhLogEntry struct {
	zl *zap.SugaredLogger
}

func NewBhLogEntry(logger *zap.SugaredLogger) *BhLogEntry {
	return &BhLogEntry{
		zl: logger,
	}
}

func (l *BhLogEntry) WithFields(fields log.Fields) log.Entry {
	l.zl.With("fields", fields)
	return l
}

func (l *BhLogEntry) WithField(name string, value interface{}) log.Entry {
	l.zl.With(name, value)
	return l
}

func (l *BhLogEntry) Debug(args ...interface{}) {
	return
}

func (l *BhLogEntry) Info(args ...interface{}) {
	return
}

func (l *BhLogEntry) Warn(args ...interface{}) {
	zlog.Log().Warn(args...)
}

func (l *BhLogEntry) Error(args ...interface{}) {
	zlog.Log().Error(args...)
}

func (l *BhLogEntry) Debugf(format string, args ...interface{}) {
	return
}

func (l *BhLogEntry) Infof(format string, args ...interface{}) {
	return
}

func (l *BhLogEntry) Warnf(format string, args ...interface{}) {
	zlog.Log().Warnf(format, args...)
}

func (l *BhLogEntry) Errorf(format string, args ...interface{}) {
	zlog.Log().Errorf(format, args...)
}
