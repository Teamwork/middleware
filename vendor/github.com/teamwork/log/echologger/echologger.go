// +build !travistest

package echologger

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/teamwork/log"

	elog "github.com/labstack/echo/log" // nolint
)

// EchoLogger extends a standard Logger struct to "support" echo's expected (and
// completely brain-dead) Logger interface.
// Most of these methods are not used internally to echo, so can be just
// scaffolding. We use our own logger for everything we control.
type EchoLogger struct {
	// This embedded interface is for the sake of satisfying the interface,
	// but is left nil so that calls to methods not specifically defined in
	// this file will panic. This should never happen, but if it does we of
	// course want to know about it.
	elog.Logger // nolint
	Entry       *log.Entry
}

// To ensure that we match the target interface
var _ elog.Logger = &EchoLogger{} // nolint

// These functions are actually used internally to echo, so we need to implement
// them.

func (l *EchoLogger) Error(i ...interface{}) {
	l.Entry.Error(errors.New("internal echo error"), i...)
}

// Fatal logs to the underlying logger, then panics
func (l *EchoLogger) Fatal(i ...interface{}) {
	msg := fmt.Sprint(i...)
	l.Entry.Print(msg)
	panic(msg)
}

// Warn logs to the underlying logger
func (l *EchoLogger) Warn(i ...interface{}) {
	l.Entry.Print(i...)
}

// Debug logs to the underlying logger
func (l *EchoLogger) Debug(i ...interface{}) {
	l.Entry.Debug(i...)
}

// Printf logs to the underlying logger
func (l *EchoLogger) Printf(fmt string, i ...interface{}) {
	l.Entry.Printf(fmt, i...)
}

const gracefulErrorPrefix = "[ERROR] "

// Output returns an io.Writer that can be used for logging. This is
// specifically tailored to handle logs from github.com/tylerb/graceful, which
// at present is only use within echo.
func (l *EchoLogger) Output() io.Writer {
	rp, wp := io.Pipe()
	go func(r io.Reader) {
		log := l.Entry
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := strings.TrimPrefix(scanner.Text(), l.Prefix()+": ")
			if strings.HasPrefix(line, gracefulErrorPrefix) {
				line = strings.TrimPrefix(line, gracefulErrorPrefix)
				log.Module("github.com/tylerb/graceful").Err(errors.New(line))
			} else {
				log.Print(line)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Module("EchoLogger").Error(err, "error reading pipe")
		}
	}(rp)
	return wp
}

// Prefix to satisfy the echo.Logger interface.
func (l *EchoLogger) Prefix() string {
	return l.Entry.GetModule()
}
