package log

import (
	"io/ioutil"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

type LoggerTester struct {
	cb func(*logrus.Entry)
}

var _ logrus.Hook = &LoggerTester{}

func (lt *LoggerTester) Fire(e *logrus.Entry) error {
	lt.cb(e)
	return nil
}

func (lt *LoggerTester) Levels() []logrus.Level {
	return logrus.AllLevels
}

func NewLoggerTester() (*Logger, *LoggerTester) {
	l := logrus.New()
	l.Out = ioutil.Discard
	h := &LoggerTester{}
	l.Hooks.Add(h)
	l.Level = logrus.DebugLevel
	return &Logger{
		logrus: l,
	}, h
}

type LoggerTest struct {
	Name    string
	Fn      func()
	Level   Level
	Message string
}

type LoggerTests []LoggerTest

func (lt LoggerTests) Run(t *testing.T, hook *LoggerTester) {
	wg := &sync.WaitGroup{}
	// First run all the tests
	for _, test := range lt {
		wg.Add(1)
		hook.cb = func(entry *logrus.Entry) {
			defer wg.Done()
			test.Check(t, entry)
		}
		test.Fn()
		wg.Wait()
	}
}

func (test *LoggerTest) Check(t *testing.T, entry *logrus.Entry) {
	if entry == nil {
		t.Errorf("Test %s: resulted in a null entry\n", test.Name)
		return
	}
	if entry.Message != test.Message {
		t.Errorf("Test %s: logged '%s', expected '%s'\n", test.Name, entry.Message, test.Message)
	}
	if Level(entry.Level) != test.Level {
		t.Errorf("Test %s: logged at level %d, expected %d\n", test.Name, entry.Level, test.Level)
	}
	data := entry.Data
	if val, ok := data["file"]; ok {
		filename := val.(string)
		data["file"] = trimFilename(filename)
		data["line"] = dummyLineNo
	}
	for _, traceName := range []string{"stackTrace", "errorStackTrace", "errorCauseStackTrace"} {
		if val, ok := data[traceName]; ok {
			stackTrace, ok := val.([]map[string]interface{})
			if !ok {
				continue
			}
			for _, frame := range stackTrace {
				filename := frame["file"].(string)
				frame["file"] = trimFilename(filename)
				frame["line"] = dummyLineNo
			}
		}
	}
}

func TestLoggerValues(t *testing.T) {
	l, hook := NewLoggerTester()
	defer SwapStdLogger(l).Unswap()
	tests := LoggerTests{
		LoggerTest{
			Name: "Global DebugLogger",
			Fn: func() {
				dl, dc := DebugLogger()
				defer dc()
				dl.Printf("Test1")
			},
			Level:   DebugLevel,
			Message: "Test1",
		},
		LoggerTest{
			Name: "Local DebugLogger",
			Fn: func() {
				dl, dc := l.DebugLogger()
				defer dc()
				dl.Printf("Test2")
			},
			Level:   DebugLevel,
			Message: "Test2",
		},
		LoggerTest{
			Name: "Entry DebugLogger",
			Fn: func() {
				dl, dc := NewEntry(l).DebugLogger()
				defer dc()
				dl.Printf("Test3")
			},
			Level:   DebugLevel,
			Message: "Test3",
		},
		LoggerTest{
			Name: "Global InfoLogger",
			Fn: func() {
				il, ic := InfoLogger()
				defer ic()
				il.Printf("Test1")
			},
			Level:   InfoLevel,
			Message: "Test1",
		},
		LoggerTest{
			Name: "Local InfoLogger",
			Fn: func() {
				il, ic := l.InfoLogger()
				defer ic()
				il.Printf("Test2")
			},
			Level:   InfoLevel,
			Message: "Test2",
		},
		LoggerTest{
			Name: "Entry InfoLogger",
			Fn: func() {
				il, ic := NewEntry(l).InfoLogger()
				defer ic()
				il.Printf("Test3")
			},
			Level:   InfoLevel,
			Message: "Test3",
		},
		LoggerTest{
			Name: "Global ErrorLogger",
			Fn: func() {
				el, ec := ErrorLogger()
				defer ec()
				el.Printf("Test1")
			},
			Level:   ErrorLevel,
			Message: "Test1",
		},
		LoggerTest{
			Name: "Local ErrorLogger",
			Fn: func() {
				el, ec := l.ErrorLogger()
				defer ec()
				el.Printf("Test2")
			},
			Level:   ErrorLevel,
			Message: "Test2",
		},
		LoggerTest{
			Name: "Entry ErrorLogger",
			Fn: func() {
				el, ec := NewEntry(l).ErrorLogger()
				defer ec()
				el.Printf("Test3")
			},
			Level:   ErrorLevel,
			Message: "Test3",
		},
	}
	tests.Run(t, hook)
}
