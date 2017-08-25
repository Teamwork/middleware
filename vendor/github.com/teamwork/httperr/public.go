package httperr

import (
	"fmt"
	"io"
)

// publicError is an error type for wrapping an error message which may contian
// technical or sensitive information, with a public-facing friendly version.
type publicError struct {
	msg string
	err error
}

var _ error = &publicError{}
var _ causer = &publicError{}

func (e *publicError) Error() string { return e.msg }
func (e *publicError) Cause() error  { return e.err }

func (e *publicError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", e.Cause())
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, e.Error())
	}
}

const silentMessage = "an error occurred"

// Silence returns a new, silent error message, which returns a generic
// "an error occurred" message. The original error message will be accessiable
// via the `diagnoser` interface. If err is nill, Silence returns nil.
func Silence(err error) error {
	if err == nil {
		return nil
	}
	return &publicError{
		msg: silentMessage,
		err: err,
	}
}

type publicWithStatus struct {
	publicError
	status int
}

// SilenceWithStatus is the same as Silence, but includes an HTTP status code.
// If err is nil, SilenceWithStatus returns nil.
func SilenceWithStatus(status int, err error) error {
	if err == nil {
		return nil
	}
	return &publicWithStatus{
		publicError: publicError{
			msg: silentMessage,
			err: err,
		},
		status: status,
	}
}
