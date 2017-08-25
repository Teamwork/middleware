package log

import (
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Entry wraps a logrus.Entry
type Entry struct {
	e      *logrus.Entry
	logger *Logger
}

// NewEntry returns a new entry for the specified Logger.
func NewEntry(l *Logger) *Entry {
	return &Entry{
		e:      logrus.NewEntry(l.logrus),
		logger: l,
	}
}

// Debug logs a message at level Debug.
func (e *Entry) Debug(args ...interface{}) {
	e.WithFields(callerFields()).e.Debug(args...)
}

// Debugf logs a message at level Debug.
func (e *Entry) Debugf(fmt string, args ...interface{}) {
	e.WithFields(callerFields()).e.Debugf(fmt, args...)
}

// Debugln logs a message at level Debug.
func (e *Entry) Debugln(args ...interface{}) {
	e.WithFields(callerFields()).e.Debugln(args...)
}

// Err logs err.Error() as a message at level Error.
func (e *Entry) Err(err error) {
	e.Error(err, err.Error())
}

// Error logs a message at level Error.
func (e *Entry) Error(err error, args ...interface{}) {
	e.WithError(err).e.Error(args...)
}

// Errorf logs a message at level Error.
func (e *Entry) Errorf(err error, fmt string, args ...interface{}) {
	e.WithError(err).e.Errorf(fmt, args...)
}

// Errorln logs a message at level Error.
func (e *Entry) Errorln(err error, args ...interface{}) {
	e.WithError(err).e.Errorln(args...)
}

// Print logs a message at level Info.
func (e *Entry) Print(args ...interface{}) {
	e.e.Print(args...)
}

// Printf logs a message at level Info.
func (e *Entry) Printf(fmt string, args ...interface{}) {
	e.e.Printf(fmt, args...)
}

// Println logs a message at level Info.
func (e *Entry) Println(args ...interface{}) {
	e.e.Println(args...)
}

// WithError returns a new entry with an error added as field.
func (e *Entry) WithError(err error) *Entry {
	if err == nil {
		err = errors.New("(nil)")
	}

	return e.WithFields(Fields{
		fieldError:        addStackTrace(err),
		fieldErrorMessage: err.Error(),
	}).WithFields(ExtractFields(err))
}

// WithHTTPRequest returns a new entry with an *http.Request added as field.
func (e *Entry) WithHTTPRequest(req *http.Request) *Entry {
	return e.WithField(fieldHTTPRequest, req)
}

// WithField returns a new entry with a single field added.
func (e *Entry) WithField(key string, value interface{}) *Entry {
	return &Entry{
		e: e.e.WithField(key, value),
	}
}

// Module adds a module field and returns a new Entry.
func (e *Entry) Module(module string) *Entry {
	return e.WithField(fieldModule, module)
}

// GetModule gets the module field.
func (e *Entry) GetModule() string {
	m, _ := e.e.Data["module"].(string)
	return m
}

// WithFields returns a new entry with a map of fields added.
func (e *Entry) WithFields(fields Fields) *Entry {
	return &Entry{
		e: e.e.WithFields(logrus.Fields(fields)),
	}
}
