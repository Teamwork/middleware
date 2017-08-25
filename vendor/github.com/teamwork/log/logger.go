package log

import (
	glog "log"

	"github.com/sirupsen/logrus"
)

// DebugLogger returns a standard library *log.Logger which logs at Debug level,
// and a function to close the underlying io.Writer when done.
func DebugLogger() (*glog.Logger, func()) {
	return NewEntry(stdLogger).DebugLogger()
}

// DebugLogger returns a standard library *log.Logger which logs at Debug level,
// and a function to close the underlying io.Writer when done.
func (l *Logger) DebugLogger() (*glog.Logger, func()) {
	return NewEntry(l).DebugLogger()
}

// DebugLogger returns a standard library *log.Logger which logs at Debug level,
// and a function to close the underlying io.Writer when done.
func (e *Entry) DebugLogger() (*glog.Logger, func()) {
	w := e.e.Logger.WriterLevel(logrus.Level(DebugLevel))
	return glog.New(w, "", 0), func() {
		e.logger.Flush()
		_ = w.Close()
	}
}

// InfoLogger returns a standard library *log.Logger which logs at Info level,
// and a function to close the underlying io.Writer when done.
func InfoLogger() (*glog.Logger, func()) {
	return NewEntry(stdLogger).InfoLogger()
}

// InfoLogger returns a standard library *log.Logger which logs at Info level,
// and a function to close the underlying io.Writer when done.
func (l *Logger) InfoLogger() (*glog.Logger, func()) {
	return NewEntry(l).InfoLogger()
}

// InfoLogger returns a standard library *log.Logger which logs at Info level,
// and a function to close the underlying io.Writer when done.
func (e *Entry) InfoLogger() (*glog.Logger, func()) {
	w := e.e.Logger.WriterLevel(logrus.Level(InfoLevel))
	return glog.New(w, "", 0), func() {
		_ = w.Close()
	}
}

// ErrorLogger returns a standard library *log.Logger which logs at Error level,
// and a function to close the underlying io.Writer when done.
func ErrorLogger() (*glog.Logger, func()) {
	return NewEntry(stdLogger).ErrorLogger()
}

// ErrorLogger returns a standard library *log.Logger which logs at Error level,
// and a function to close the underlying io.Writer when done.
func (l *Logger) ErrorLogger() (*glog.Logger, func()) {
	return NewEntry(l).ErrorLogger()
}

// ErrorLogger returns a standard library *log.Logger which logs at Error level,
// and a function to close the underlying io.Writer when done.
func (e *Entry) ErrorLogger() (*glog.Logger, func()) {
	w := e.e.Logger.WriterLevel(logrus.Level(ErrorLevel))
	return glog.New(w, "", 0), func() {
		_ = w.Close()
	}
}
