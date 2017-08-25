package log

import (
	"errors"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

func TestLocalLogger(t *testing.T) {
	l, hook := NewTestLogger()
	err := errors.New("Test")
	errWrapped := pkgerrors.Wrap(err, "Secondary error")
	errWithStack := pkgerrors.New("Stack Error")
	errWrappedWithStack := pkgerrors.Wrap(errWithStack, "Secondary error")
	errDoubleWrapped := pkgerrors.Wrap(errWrapped, "Terciary error")
	stack := dummyStack()

	tests := LogTests([]LogTest{
		LogTest{
			Name:    "Local Debug",
			Fn:      func() { l.Debug("x") },
			Level:   DebugLevel,
			Message: "x",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestLocalLogger.func1",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Local Debugf",
			Fn:      func() { l.Debugf("%s", "y") },
			Level:   DebugLevel,
			Message: "y",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestLocalLogger.func2",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Local Debugln",
			Fn:      func() { l.Debugln("z") },
			Level:   DebugLevel,
			Message: "z",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestLocalLogger.func3",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Local Print",
			Fn:      func() { l.Print("x") },
			Level:   InfoLevel,
			Message: "x",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Local Printf",
			Fn:      func() { l.Printf("%s", "y") },
			Level:   InfoLevel,
			Message: "y",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Local Println",
			Fn:      func() { l.Println("z") },
			Level:   InfoLevel,
			Message: "z",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Local Err",
			Fn:      func() { l.Err(err) },
			Level:   ErrorLevel,
			Message: "Test",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Local Error",
			Fn:      func() { l.Error(err, "x") },
			Level:   ErrorLevel,
			Message: "x",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Local Error with Cause",
			Fn:      func() { l.Error(errWrapped, "x") },
			Level:   ErrorLevel,
			Message: "x",
			Fields: map[string]interface{}{
				fieldErrorMessage: errWrapped.Error(),
				fieldError:        errWrapped,
			},
		},
		LogTest{
			Name:    "Local Error with Cause and Stack",
			Fn:      func() { l.Error(errWrappedWithStack, "x") },
			Level:   ErrorLevel,
			Message: "x",
			Fields: map[string]interface{}{
				fieldErrorMessage: errWrappedWithStack.Error(),
				fieldError:        errWrappedWithStack,
			},
		},
		LogTest{
			Name:    "Local Double-Wrapped Error with Cause",
			Fn:      func() { l.Error(errDoubleWrapped, "x") },
			Level:   ErrorLevel,
			Message: "x",
			Fields: map[string]interface{}{
				fieldErrorMessage: errDoubleWrapped.Error(),
				fieldError:        errDoubleWrapped,
			},
		},
		LogTest{
			Name:    "Local Errorf",
			Fn:      func() { l.Errorf(err, "%s", "y") },
			Level:   ErrorLevel,
			Message: "y",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Local Errorln",
			Fn:      func() { l.Errorln(err, "z") },
			Level:   ErrorLevel,
			Message: "z",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Local WithError",
			Fn:      func() { l.WithError(err).Print("a") },
			Level:   InfoLevel,
			Message: "a",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Local WithField",
			Fn:      func() { l.WithField("foo", "bar").Print("b") },
			Level:   InfoLevel,
			Message: "b",
			Fields:  map[string]interface{}{"foo": "bar"},
		},
		LogTest{
			Name:    "Local WithFields",
			Fn:      func() { l.WithFields(Fields(map[string]interface{}{"moo": "quack"})).Print("c") },
			Level:   InfoLevel,
			Message: "c",
			Fields:  map[string]interface{}{"moo": "quack"},
		},
	})
	tests.Run(t, hook)
}

func TestGlobalLogger(t *testing.T) {
	l, hook := NewTestLogger()
	defer SwapStdLogger(l).Unswap()
	err := errors.New("Test")
	stack := dummyStack()

	tests := LogTests([]LogTest{
		LogTest{
			Name:    "Global Debug",
			Fn:      func() { Debug("x") },
			Level:   DebugLevel,
			Message: "x",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestGlobalLogger.func1",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Global Debugf",
			Fn:      func() { Debugf("%s", "y") },
			Level:   DebugLevel,
			Message: "y",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestGlobalLogger.func2",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Global Debugln",
			Fn:      func() { Debugln("z") },
			Level:   DebugLevel,
			Message: "z",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestGlobalLogger.func3",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Global Print",
			Fn:      func() { Print("x") },
			Level:   InfoLevel,
			Message: "x",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Global Printf",
			Fn:      func() { Printf("%s", "y") },
			Level:   InfoLevel,
			Message: "y",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Global Println",
			Fn:      func() { Println("z") },
			Level:   InfoLevel,
			Message: "z",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Global Err",
			Fn:      func() { Err(err) },
			Level:   ErrorLevel,
			Message: "Test",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Global Error",
			Fn:      func() { Error(err, "x") },
			Level:   ErrorLevel,
			Message: "x",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Global Errorf",
			Fn:      func() { Errorf(err, "%s", "y") },
			Level:   ErrorLevel,
			Message: "y",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Global Errorln",
			Fn:      func() { Errorln(err, "z") },
			Level:   ErrorLevel,
			Message: "z",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Global WithError",
			Fn:      func() { WithError(err).Print("a") },
			Level:   InfoLevel,
			Message: "a",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Global WithField",
			Fn:      func() { WithField("foo", "bar").Print("b") },
			Level:   InfoLevel,
			Message: "b",
			Fields:  map[string]interface{}{"foo": "bar"},
		},
		LogTest{
			Name:    "Global WithFields",
			Fn:      func() { WithFields(Fields(map[string]interface{}{"moo": "quack"})).Print("c") },
			Level:   InfoLevel,
			Message: "c",
			Fields:  map[string]interface{}{"moo": "quack"},
		},
	})
	tests.Run(t, hook)
}

func TestEntryLogger(t *testing.T) {
	l, hook := NewTestLogger()
	e := NewEntry(l)
	err := errors.New("Test")
	stack := dummyStack()

	tests := LogTests([]LogTest{
		LogTest{
			Name:    "Entry Debug",
			Fn:      func() { e.Debug("x") },
			Level:   DebugLevel,
			Message: "x",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestEntryLogger.func1",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Entry Debugf",
			Fn:      func() { e.Debugf("%s", "y") },
			Level:   DebugLevel,
			Message: "y",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestEntryLogger.func2",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Entry Debugln",
			Fn:      func() { e.Debugln("z") },
			Level:   DebugLevel,
			Message: "z",
			Fields: map[string]interface{}{
				"file":     "teamwork/log/log_test.go",
				"function": "github.com/teamwork/log.TestEntryLogger.func3",
				"line":     dummyLineNo,
			},
		},
		LogTest{
			Name:    "Entry Print",
			Fn:      func() { e.Print("x") },
			Level:   InfoLevel,
			Message: "x",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Entry Printf",
			Fn:      func() { e.Printf("%s", "y") },
			Level:   InfoLevel,
			Message: "y",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Entry Println",
			Fn:      func() { e.Println("z") },
			Level:   InfoLevel,
			Message: "z",
			Fields:  map[string]interface{}{},
		},
		LogTest{
			Name:    "Entry Err",
			Fn:      func() { e.Err(err) },
			Level:   ErrorLevel,
			Message: "Test",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Entry Error",
			Fn:      func() { e.Error(err, "x") },
			Level:   ErrorLevel,
			Message: "x",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Entry Errorf",
			Fn:      func() { e.Errorf(err, "%s", "y") },
			Level:   ErrorLevel,
			Message: "y",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Entry Errorln",
			Fn:      func() { e.Errorln(err, "z") },
			Level:   ErrorLevel,
			Message: "z",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Entry WithError",
			Fn:      func() { e.WithError(err).Print("a") },
			Level:   InfoLevel,
			Message: "a",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
			},
		},
		LogTest{
			Name:    "Entry WithField",
			Fn:      func() { e.WithField("foo", "bar").Print("b") },
			Level:   InfoLevel,
			Message: "b",
			Fields:  map[string]interface{}{"foo": "bar"},
		},
		LogTest{
			Name:    "Entry WithFields",
			Fn:      func() { e.WithFields(Fields(map[string]interface{}{"moo": "quack"})).Print("c") },
			Level:   InfoLevel,
			Message: "c",
			Fields:  map[string]interface{}{"moo": "quack"},
		},
	})
	tests.Run(t, hook)
}
