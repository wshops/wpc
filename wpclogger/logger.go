package wpclogger

import (
	"github.com/apache/pulsar-client-go/pulsar/log"
	"go.uber.org/zap"
)

var zlog *ZapLog

type ZapLog struct {
	sl *zap.SugaredLogger
}

func SetLogger(logger *zap.SugaredLogger) {
	zlog = &ZapLog{
		sl: logger,
	}
}

type WpcLogger struct {
	le *LogEntry
}

func NewWpcLogger() *WpcLogger {
	return &WpcLogger{
		le: NewLogEntry(),
	}
}

func (w *WpcLogger) SubLogger(fields log.Fields) log.Logger {
	zlog.sl.With(fields)
	return w
}

func (w *WpcLogger) WithFields(fields log.Fields) log.Entry {
	return w.le.WithFields(fields)
}

func (w *WpcLogger) WithField(name string, value interface{}) log.Entry {
	return w.le.WithField(name, value)
}

func (w *WpcLogger) WithError(err error) log.Entry {
	return w.le.WithField("error", err)
}

func (w *WpcLogger) Debug(args ...interface{}) {
	zlog.sl.Debug(args...)
}

func (w *WpcLogger) Info(args ...interface{}) {
	zlog.sl.Info(args...)
}

func (w *WpcLogger) Warn(args ...interface{}) {
	zlog.sl.Warn(args...)
}

func (w *WpcLogger) Error(args ...interface{}) {
	zlog.sl.Error(args...)
}

func (w *WpcLogger) Debugf(format string, args ...interface{}) {
	zlog.sl.Debugf(format, args...)
}

func (w *WpcLogger) Infof(format string, args ...interface{}) {
	zlog.sl.Infof(format, args...)
}

func (w *WpcLogger) Warnf(format string, args ...interface{}) {
	zlog.sl.Warnf(format, args...)
}

func (w *WpcLogger) Errorf(format string, args ...interface{}) {
	zlog.sl.Errorf(format, args...)
}

type LogEntry struct {
}

func NewLogEntry() *LogEntry {
	return &LogEntry{}
}

func (l *LogEntry) WithFields(fields log.Fields) log.Entry {
	zlog.sl.With(fields)
	return l
}

func (l *LogEntry) WithField(name string, value interface{}) log.Entry {
	zlog.sl.With(name, value)
	return l
}

func (l *LogEntry) Debug(args ...interface{}) {
	zlog.sl.Debug(args...)
}

func (l *LogEntry) Info(args ...interface{}) {
	zlog.sl.Info(args...)
}

func (l *LogEntry) Warn(args ...interface{}) {
	zlog.sl.Warn(args...)
}

func (l *LogEntry) Error(args ...interface{}) {
	zlog.sl.Error(args...)
}

func (l *LogEntry) Debugf(format string, args ...interface{}) {
	zlog.sl.Debugf(format, args...)
}

func (l *LogEntry) Infof(format string, args ...interface{}) {
	zlog.sl.Infof(format, args...)
}

func (l *LogEntry) Warnf(format string, args ...interface{}) {
	zlog.sl.Warnf(format, args...)
}

func (l *LogEntry) Errorf(format string, args ...interface{}) {
	zlog.sl.Errorf(format, args...)
}
