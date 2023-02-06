package wpclogger

import (
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/gookit/slog"
)

type WpcLogger struct {
	le *LogEntry
}

func NewWpcLogger() *WpcLogger {
	return &WpcLogger{
		le: NewLogEntry(),
	}
}

func (w *WpcLogger) SubLogger(fields log.Fields) log.Logger {
	slog.WithFields(slog.M(fields))
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
	slog.Debug(args...)
}

func (w *WpcLogger) Info(args ...interface{}) {
	slog.Info(args...)
}

func (w *WpcLogger) Warn(args ...interface{}) {
	slog.Warn(args...)
}

func (w *WpcLogger) Error(args ...interface{}) {
	slog.Error(args...)
}

func (w *WpcLogger) Debugf(format string, args ...interface{}) {
	slog.Debugf(format, args...)
}

func (w *WpcLogger) Infof(format string, args ...interface{}) {
	slog.Infof(format, args...)
}

func (w *WpcLogger) Warnf(format string, args ...interface{}) {
	slog.Warnf(format, args...)
}

func (w *WpcLogger) Errorf(format string, args ...interface{}) {
	slog.Errorf(format, args...)
}

type LogEntry struct {
}

func NewLogEntry() *LogEntry {
	return &LogEntry{}
}

func (l *LogEntry) WithFields(fields log.Fields) log.Entry {
	slog.WithFields(slog.M(fields))
	return l
}

func (l *LogEntry) WithField(name string, value interface{}) log.Entry {
	slog.WithFields(map[string]any{name: value})
	return l
}

func (l *LogEntry) Debug(args ...interface{}) {
	slog.Debug(args...)
}

func (l *LogEntry) Info(args ...interface{}) {
	slog.Info(args...)
}

func (l *LogEntry) Warn(args ...interface{}) {
	slog.Warn(args...)
}

func (l *LogEntry) Error(args ...interface{}) {
	slog.Error(args...)
}

func (l *LogEntry) Debugf(format string, args ...interface{}) {
	slog.Debugf(format, args...)
}

func (l *LogEntry) Infof(format string, args ...interface{}) {
	slog.Infof(format, args...)
}

func (l *LogEntry) Warnf(format string, args ...interface{}) {
	slog.Warnf(format, args...)
}

func (l *LogEntry) Errorf(format string, args ...interface{}) {
	slog.Errorf(format, args...)
}
