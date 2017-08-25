package log

import (
	"io"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

var stdLogger = New()

// SetStandardLogger sets the global, standard logger
func SetStandardLogger(l *Logger) {
	stdLogger = l
}

// StandardLogger returns the standard logger
func StandardLogger() *Logger {
	return stdLogger
}

// New creates a new logger.
func New() *Logger {
	l := logrus.New()
	l.Formatter = &twFormatter{}
	return &Logger{
		logrus: l,
	}
}

// NewWithLogrus creates a new logger with the given logrus Logger.
func NewWithLogrus(l *logrus.Logger) *Logger {
	return &Logger{
		logrus: l,
	}
}

// Logger is a wrapper around a logrus.Logger
type Logger struct {
	logrus   *logrus.Logger
	flushers []flusher
}

type flusher interface {
	Flush()
}

// Flush ensures that the underlying hooks' queues are flushed.
func (l *Logger) Flush() {
	if l.flushers == nil {
		return
	}
	wg := &sync.WaitGroup{}
	for _, f := range l.flushers {
		wg.Add(1)
		go func(f flusher) {
			defer wg.Done()
			f.Flush()
		}(f)
	}
	wg.Wait()
}

// Flush ensures that the standard logger's hooks' queues are flushed.
func Flush() {
	stdLogger.Flush()
}

// SetFormatter sets the Logger's formatter
func (l *Logger) SetFormatter(f logrus.Formatter) {
	l.logrus.Formatter = f
}

// SetOutput sets the Logger's output writer
func (l *Logger) SetOutput(out io.Writer) {
	l.logrus.Out = out
}

// AddHook adds a hook to the logger
func (l *Logger) AddHook(hook logrus.Hook) {
	l.logrus.Hooks.Add(hook)
}

// Debug logs a message at level Debug.
func (l *Logger) Debug(args ...interface{}) {
	NewEntry(l).Debug(args...)
}

// Debug logs a message at Debug Info on the standard logger.
func Debug(args ...interface{}) {
	stdLogger.Debug(args...)
}

// Debugf logs a message at level Debug.
func (l *Logger) Debugf(fmt string, args ...interface{}) {
	NewEntry(l).Debugf(fmt, args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(fmt string, args ...interface{}) {
	NewEntry(stdLogger).Debugf(fmt, args...)
}

// Debugln logs a message at level Debug.
func (l *Logger) Debugln(args ...interface{}) {
	NewEntry(l).Debugln(args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	NewEntry(stdLogger).Debugln(args...)
}

// Print logs a message at level Info.
func (l *Logger) Print(args ...interface{}) {
	NewEntry(l).Print(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	NewEntry(stdLogger).Print(args...)
}

// Printf logs a message at level Info.
func (l *Logger) Printf(fmt string, args ...interface{}) {
	l.logrus.Printf(fmt, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(fmt string, args ...interface{}) {
	stdLogger.Printf(fmt, args...)
}

// Println logs a message at level Info.
func (l *Logger) Println(args ...interface{}) {
	l.logrus.Println(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(args ...interface{}) {
	stdLogger.Println(args...)
}

// Err logs err.Error() as a message at level Error.
func (l *Logger) Err(err error) {
	NewEntry(l).Err(err)
}

// Err logs err.Error() as a message at level Error.
func Err(err error) {
	NewEntry(stdLogger).Err(err)
}

// Error logs a message at level Error.
func (l *Logger) Error(err error, args ...interface{}) {
	NewEntry(l).Error(err, args...)
}

// Error logs a message at level Error on the standard logger.
func Error(err error, args ...interface{}) {
	NewEntry(stdLogger).Error(err, args...)
}

// Errorf logs a message at level Error.
func (l *Logger) Errorf(err error, fmt string, args ...interface{}) {
	NewEntry(l).Errorf(err, fmt, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(err error, fmt string, args ...interface{}) {
	NewEntry(stdLogger).Errorf(err, fmt, args...)
}

// Errorln logs a message at level Error.
func (l *Logger) Errorln(err error, args ...interface{}) {
	NewEntry(l).Errorln(err, args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(err error, args ...interface{}) {
	stdLogger.Errorln(err, args...)
}

// WithError adds an error as single field to a the standard logger.
func WithError(err error) *Entry {
	return NewEntry(stdLogger).WithError(err)
}

// WithError adds an error as single field to a new log Entry.
func (l *Logger) WithError(err error) *Entry {
	return NewEntry(l).WithError(err)
}

// WithHTTPRequest adds an *http.Request field to the standard logger.
func WithHTTPRequest(req *http.Request) *Entry {
	return NewEntry(stdLogger).WithHTTPRequest(req)
}

// WithHTTPRequest adds an *http.Request field to the standard logger.
func (l *Logger) WithHTTPRequest(req *http.Request) *Entry {
	return NewEntry(l).WithHTTPRequest(req)
}

// WithField adds a field to the standard logger and returns a new Entry.
func WithField(key string, value interface{}) *Entry {
	return NewEntry(stdLogger).WithField(key, value)
}

// WithField adds a field to the log entry, note that it doesn't log until you
// call Debug, Print, or Error. It only creates a log entry. If you want
// multiple fields, use `WithFields`.
func (l *Logger) WithField(key string, value interface{}) *Entry {
	return NewEntry(l).WithField(key, value)
}

// Module adds a module field to the standard logger and returns a new Entry.
func Module(module string) *Entry {
	return NewEntry(stdLogger).Module(module)
}

// Module adds a module field and returns a new Entry.
func (l *Logger) Module(module string) *Entry {
	return NewEntry(l).Module(module)
}

// WithFields adds a map of fields to the standard logger and returns a new Entry.
func WithFields(fields Fields) *Entry {
	return stdLogger.WithFields(fields)
}

// WithFields adds a map of fields to the log entry, note that it doesn't log
// until you call Debug, Print, or Error. It only creates a log entry. If you
// want multiple fields, use `WithFields`.
func (l *Logger) WithFields(fields Fields) *Entry {
	e := l.logrus.WithFields(logrus.Fields(fields))
	return &Entry{e: e}
}
