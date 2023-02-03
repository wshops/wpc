package wpclogger

import (
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/gookit/slog"
)

type WpcLogger struct {
	lg *slog.Logger
	le *LogEntry
}

func NewWpcLogger() *WpcLogger {
	l := slog.NewStdLogger().Logger
	return &WpcLogger{
		lg: l,
		le: NewLogEntry(l),
	}
}

func (w *WpcLogger) SubLogger(fields log.Fields) log.Logger {
	w.lg.WithFields(slog.M(fields))
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
	w.lg.Debug(args...)
}

func (w *WpcLogger) Info(args ...interface{}) {
	w.lg.Info(args...)
}

func (w *WpcLogger) Warn(args ...interface{}) {
	w.lg.Warn(args...)
}

func (w *WpcLogger) Error(args ...interface{}) {
	w.lg.Error(args...)
}

func (w *WpcLogger) Debugf(format string, args ...interface{}) {
	w.lg.Debugf(format, args...)
}

func (w *WpcLogger) Infof(format string, args ...interface{}) {
	w.lg.Infof(format, args...)
}

func (w *WpcLogger) Warnf(format string, args ...interface{}) {
	w.lg.Warnf(format, args...)
}

func (w *WpcLogger) Errorf(format string, args ...interface{}) {
	w.lg.Errorf(format, args...)
}

type LogEntry struct {
	lg *slog.Logger
}

func NewLogEntry(l *slog.Logger) *LogEntry {
	return &LogEntry{lg: l}
}

func (l *LogEntry) WithFields(fields log.Fields) log.Entry {
	l.lg.WithFields(slog.M(fields))
	return l
}

func (l *LogEntry) WithField(name string, value interface{}) log.Entry {
	l.lg.WithField(name, value)
	return l
}

func (l *LogEntry) Debug(args ...interface{}) {
	l.lg.Debug(args...)
}

func (l *LogEntry) Info(args ...interface{}) {
	l.lg.Info(args...)
}

func (l *LogEntry) Warn(args ...interface{}) {
	l.lg.Warn(args...)
}

func (l *LogEntry) Error(args ...interface{}) {
	l.lg.Error(args...)
}

func (l *LogEntry) Debugf(format string, args ...interface{}) {
	l.lg.Debugf(format, args...)
}

func (l *LogEntry) Infof(format string, args ...interface{}) {
	l.lg.Infof(format, args...)
}

func (l *LogEntry) Warnf(format string, args ...interface{}) {
	l.lg.Warnf(format, args...)
}

func (l *LogEntry) Errorf(format string, args ...interface{}) {
	l.lg.Errorf(format, args...)
}
