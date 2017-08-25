package log

import "github.com/sirupsen/logrus"

// Level represents a supported logging level (copied from logrus.Level)
type Level uint8

const (
	// ErrorLevel is self-explanatory
	ErrorLevel = Level(logrus.ErrorLevel)
	// InfoLevel is self-explanatory
	InfoLevel = Level(logrus.InfoLevel)
	// DebugLevel is self-explanatory
	DebugLevel = Level(logrus.DebugLevel)
)

const (
	fieldError        = "error"
	fieldErrorMessage = "error_message"
	fieldModule       = "module"
	fieldHTTPRequest  = "http_request"
)

// Fields type, used to pass to `WithFields`. An exact copy of logrus.Fields
type Fields map[string]interface{}

// Convert the Level to a string. E.g. InfoLevel becomes "info".
func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case ErrorLevel:
		return "error"
	}

	return "unknown"
}
