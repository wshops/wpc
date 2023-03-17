package wpcl

import (
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/wshops/zlog"
	"go.uber.org/zap"
)

type WpcLogger struct {
	le     *LogEntry
	zapLog *zap.SugaredLogger
}

func NewFullInfoLogger(zapLogger *zap.SugaredLogger) *WpcLogger {
	return &WpcLogger{
		le:     NewLogEntry(zapLogger),
		zapLog: zapLogger,
	}
}

func (w *WpcLogger) SubLogger(fields log.Fields) log.Logger {
	w.zapLog.With("fields", fields)
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
	w.zapLog.Debug(args...)
}

func (w *WpcLogger) Info(args ...interface{}) {
	w.zapLog.Info(args...)
}

func (w *WpcLogger) Warn(args ...interface{}) {
	w.zapLog.Warn(args...)
}

func (w *WpcLogger) Error(args ...interface{}) {
	w.zapLog.Error(args...)
}

func (w *WpcLogger) Debugf(format string, args ...interface{}) {
	w.zapLog.Debugf(format, args...)
}

func (w *WpcLogger) Infof(format string, args ...interface{}) {
	w.zapLog.Infof(format, args...)
}

func (w *WpcLogger) Warnf(format string, args ...interface{}) {
	w.zapLog.Warnf(format, args...)
}

func (w *WpcLogger) Errorf(format string, args ...interface{}) {
	zlog.Log().Errorf(format, args...)
}

type LogEntry struct {
	zapLog *zap.SugaredLogger
}

func NewLogEntry(zapLogger *zap.SugaredLogger) *LogEntry {
	return &LogEntry{
		zapLog: zapLogger,
	}
}

func (l *LogEntry) WithFields(fields log.Fields) log.Entry {
	l.zapLog.With("fields", fields)
	return l
}

func (l *LogEntry) WithField(name string, value interface{}) log.Entry {
	l.zapLog.With(name, value)
	return l
}

func (l *LogEntry) Debug(args ...interface{}) {
	l.zapLog.Debug(args...)
}

func (l *LogEntry) Info(args ...interface{}) {
	l.zapLog.Info(args...)
}

func (l *LogEntry) Warn(args ...interface{}) {
	l.zapLog.Warn(args...)
}

func (l *LogEntry) Error(args ...interface{}) {
	l.zapLog.Error(args...)
}

func (l *LogEntry) Debugf(format string, args ...interface{}) {
	l.zapLog.Debugf(format, args...)
}

func (l *LogEntry) Infof(format string, args ...interface{}) {
	l.zapLog.Infof(format, args...)
}

func (l *LogEntry) Warnf(format string, args ...interface{}) {
	l.zapLog.Warnf(format, args...)
}

func (l *LogEntry) Errorf(format string, args ...interface{}) {
	l.zapLog.Errorf(format, args...)
}
