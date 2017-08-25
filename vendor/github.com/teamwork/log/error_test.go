package log

import (
	goerr "errors"
	"testing"
)

func TestNilError(t *testing.T) {
	if ErrorWithFields(nil, Fields{}) != nil {
		t.Error("ErrorWithFields should return nil if passed a nil error")
	}
}

func TestError(t *testing.T) {
	l, hook := NewTestLogger()
	cause := goerr.New("go err")
	err := ErrorWithFields(cause, Fields{
		"foo": 10,
		"bar": "baz",
	})
	err2 := l.Module("foo").ErrorWithFields(cause, Fields{
		"foo": 10,
		"bar": "baz",
	})
	stack := dummyStack()

	tests := LogTests{
		LogTest{
			Name:    "ErrorWithFields",
			Fn:      func() { l.WithError(err).Print("hello") },
			Level:   InfoLevel,
			Message: "hello",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldError: &withStack{
					err:   err,
					stack: stack,
				},
				"foo": 10,
				"bar": "baz",
			},
		},
		LogTest{
			Name:    "entry.ErrorWithFields",
			Fn:      func() { l.WithError(err2).Print("hello") },
			Level:   InfoLevel,
			Message: "hello",
			Fields: map[string]interface{}{
				fieldErrorMessage: err.Error(),
				fieldModule:       "foo",
				fieldError: &withStack{
					err:   err2,
					stack: stack,
				},
				"foo": 10,
				"bar": "baz",
			},
		},
	}
	tests.Run(t, hook)
}

func TestErrorNil(t *testing.T) {
	// Just make sure it doesn't panic.
	Error(nil, "foo")
	Errorf(nil, "foo")
}
