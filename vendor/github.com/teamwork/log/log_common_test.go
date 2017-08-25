package log

import (
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/teamwork/log/testhook"
	"github.com/teamwork/test/diff"
)

const (
	dummyLineNo = 999 // For test output consistency
	dummyPtr    = 0xffffff
)

func dummyStack() errors.StackTrace {
	stack := errors.New("foo").(stackTracer).StackTrace()
	dummyStack := make(errors.StackTrace, len(stack))
	for i := range dummyStack {
		dummyStack[i] = dummyPtr
	}
	return dummyStack
}

type LogTest struct {
	Name    string
	Fn      func()
	Level   Level
	Message string
	Fields  logrus.Fields
}

type LogTests []LogTest

func (lt LogTests) Run(t *testing.T, hook *testhook.Hook) {
	for _, test := range lt {
		test.Run(t, hook)
	}

}

type LogSwapper struct {
	logger *Logger
}

var globalLock sync.Mutex

func SwapStdLogger(l *Logger) *LogSwapper {
	globalLock.Lock()
	oldStd := stdLogger
	stdLogger = l
	return &LogSwapper{oldStd}
}

func (ls *LogSwapper) Unswap() {
	stdLogger = ls.logger
	globalLock.Unlock()
}

// NewTestLogger returns a new logger configured for test use
func NewTestLogger() (*Logger, *testhook.Hook) {
	l, h := testhook.NewNullLogger()
	l.Level = logrus.DebugLevel
	return &Logger{
		logrus: l,
	}, h
}

func (test *LogTest) Run(t *testing.T, hook *testhook.Hook) {
	hook.Reset()
	t.Run(test.Name, func(t *testing.T) {
		test.Fn()
		if entries := len(hook.AllEntries()); entries != 1 {
			t.Errorf("Expected 1 entry, got %d\n", entries)
		}
		entry := hook.LastEntry()
		if entry == nil {
			t.Errorf("Resulted in a null entry")
			return
		}
		if entry.Message != test.Message {
			t.Errorf("Logged '%s', expected '%s'", entry.Message, test.Message)
		}
		if Level(entry.Level) != test.Level {
			t.Errorf("Logged at level %d, expected %d", entry.Level, test.Level)
		}
		data := entry.Data
		if val, ok := data["file"]; ok {
			filename := val.(string)
			data["file"] = trimFilename(filename)
			data["line"] = dummyLineNo
		}
		if err, ok := data[fieldError].(*withStack); ok {
			newStack := make(errors.StackTrace, 0, len(err.stack))
			for _, pc := range err.stack {
				fn := runtime.FuncForPC(uintptr(pc))
				file, _ := fn.FileLine(uintptr(pc))
				if !strings.Contains(file, "/_test/_obj_test/") { // Skip testing frames
					newStack = append(newStack, dummyPtr)
				}
			}
			err.stack = newStack
		}
		if d := diff.Diff(test.Fields, entry.Data); d != "" {
			t.Errorf("Fields differ:\n%s\n", d)
		}
	})
}

// This is used to normalize file names in stack traces, so tests are deterministic
func trimFilename(filename string) string {
	parts := strings.Split(filename, "/")
	return strings.Join(parts[len(parts)-3:], "/")
}
