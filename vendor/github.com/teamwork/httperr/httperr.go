package httperr

import (
	"fmt"

	"github.com/pkg/errors"
)

// statusCoder is the main interface to errors in this package.
type statusCoder interface {
	StatusCode() int
}

// causer is a local version of github.com/pkg/errors.causer
type causer interface {
	Cause() error
}

// stackTracer is a local version of github.com/pkg/errors.stackTracer
type stackTracer interface {
	StackTrace() errors.StackTrace
}

// withStatus wraps a standard error and an HTTP status code.
type withStatus struct {
	error
	status int
	*stack
}

var _ error = &withStatus{}
var _ statusCoder = &withStatus{}
var _ causer = &withStatus{}
var _ stackTracer = &withStatus{}

func (e *withStatus) Cause() error    { return e.error }
func (e *withStatus) StatusCode() int { return e.status }

// New returns a new error message with stack trace, tied to the provided HTTP
// status code.
func New(status int, msg string) error {
	return &withStatus{
		error:  errors.New(msg),
		status: status,
		stack:  callers(),
	}
}

// Errorf returns a new error message with stack trace, tied to the provided
// HTTP status code.
func Errorf(status int, format string, args ...interface{}) error {
	return &withStatus{
		error:  fmt.Errorf(format, args...),
		status: status,
		stack:  callers(),
	}
}

// WithStatus wraps an existing error with the provided HTTP status code and
// stack trace. If err is nil, WithSatus returns nil.
func WithStatus(status int, err error) error {
	if err == nil {
		return nil
	}
	return &withStatus{
		error:  err,
		status: status,
		stack:  callers(),
	}
}

type wrapped struct {
	msg    string
	status int
	error
	*stack
}

// StatusCode extracts the highest-level status code from the error, or the
// errors it wraps.
func StatusCode(err error) int {
	for {
		if sc, ok := err.(statusCoder); ok {
			return sc.StatusCode()
		}
		if cause, ok := err.(causer); ok {
			err = cause.Cause()
		} else {
			break
		}
	}
	return 0
}

var _ error = &wrapped{}
var _ stackTracer = &wrapped{}
var _ statusCoder = &wrapped{}
var _ causer = &wrapped{}

func (e *wrapped) Error() string   { return e.msg }
func (e *wrapped) Cause() error    { return e.error }
func (e *wrapped) StatusCode() int { return e.status }

// Wrap returns an error annotating err with a stack trace, HTTP status code,
// and the supplied message. If err is nil, Wrap returns nil.
func Wrap(status int, err error, msg string) error {
	if err == nil {
		return nil
	}
	return &wrapped{
		msg:    msg,
		status: status,
		error:  err,
		stack:  callers(),
	}
}

// Wrapf returns an error annotating err with a stack trace, HTTP status code,
// and the format specifier. If err is nil, Wrapf returns nil.
func Wrapf(status int, err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &wrapped{
		msg:    fmt.Sprintf(msg, args...),
		status: status,
		error:  err,
		stack:  callers(),
	}
}
