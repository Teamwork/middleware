// +build !travistest

package echologger

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/teamwork/log"
	"github.com/teamwork/log/testhook"
)

// NewTestLogger returns a new logger configured for test use
func NewTestLogger() (*log.Logger, *testhook.Hook) {
	l, h := testhook.NewNullLogger()
	l.Level = logrus.DebugLevel
	return log.NewWithLogrus(l), h
}

func TestEchoLogger(t *testing.T) {
	l, _ := NewTestLogger()
	elog := &EchoLogger{Entry: l.Module("echo")}
	err := testUnimplemented(func() {
		elog.Info("this should fail") // nolint
	})
	if err == nil {
		t.Errorf("Expected error trying to log to Info")
	}
}

func testUnimplemented(fn func()) (err error) {
	defer func() {
		r := recover()
		err = r.(error)
	}()
	fn()
	return nil
}

func TestOutput(t *testing.T) {
	l, hook := NewTestLogger()
	elog := &EchoLogger{Entry: l.Module("echo")}
	w := elog.Output()
	fmt.Fprintln(w, "[ERROR] Test error")
	fmt.Fprintln(w, "Non-test error")
	// This is just to guarnatee that the first two logs have been processed
	// before we test them, to avoid race conditions.
	fmt.Fprintln(w, "discarded")
	expectedMessages := []string{"Test error", "Non-test error"}
	expectedLevels := []log.Level{log.ErrorLevel, log.InfoLevel}
	entries := hook.AllEntries()
	for i := 0; i < 2; i++ {
		entry := entries[i]
		if entry.Message != expectedMessages[i] {
			t.Errorf("Expected message '%s', got '%s'\n", expectedMessages[i], entry.Message)
		}
		if entry.Level != logrus.Level(expectedLevels[i]) {
			t.Errorf("Expected level %s, got %s\n", expectedLevels[i], entry.Level)
		}
	}
}
