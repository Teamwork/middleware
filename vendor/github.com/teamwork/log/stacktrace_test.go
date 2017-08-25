package log

import (
	"fmt"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

func TestNilStacktraceErr(t *testing.T) {
	err := &withStack{err: nil}
	if err.Error() != "" {
		t.Error("nil error should produce an empty stacktrace")
	}
}

func TestErrorStack(t *testing.T) {
	err := pkgerrors.New("test")
	stack := errorStack(err.(stackTracer))
	result := fmt.Sprintf("%+v", stack)
	for _, want := range []string{"teamwork/log/stacktrace_test.go",
		"Function:github.com/teamwork/log.TestErrorStack"} {
		if !strings.Contains(result, want) {
			t.Errorf("Expected stack trace to contain '%s':\n%s\n", want, result)
		}
	}
}
