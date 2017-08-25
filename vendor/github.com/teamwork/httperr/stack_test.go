package httperr

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func testFormatRegexp(t *testing.T, n int, arg interface{}, format, want string) {
	got := fmt.Sprintf(format, arg)
	gotLines := strings.SplitN(got, "\n", -1)
	wantLines := strings.SplitN(want, "\n", -1)

	if len(wantLines) > len(gotLines) {
		t.Errorf("test %d: wantLines(%d) > gotLines(%d):\n got: %q\nwant: %q", n+1, len(wantLines), len(gotLines), got, want)
		return
	}

	for i, w := range wantLines {
		match, err := regexp.MatchString(w, gotLines[i])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("test %d: line %d: fmt.Sprintf(%q, err):\n got: %q\nwant: %q", n+1, i+1, format, got, want)
		}
	}
}

func stackTrace() errors.StackTrace {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(1, pcs[:])
	var st stack = pcs[0:n]
	return st.StackTrace()
}

func TestStackTraceFormat(t *testing.T) {
	tests := []struct {
		errors.StackTrace
		format string
		want   string
	}{{
		stackTrace()[:2],
		"%s",
		`\[stack_test.go stack_test.go\]`,
	}, {
		stackTrace()[:2],
		"%v",
		`\[stack_test.go:37 stack_test.go:52\]`,
	}, {
		stackTrace()[:2],
		"%+v",
		"\n" +
			"github.com/teamwork/httperr.stackTrace\n" +
			"\t.+/github.com/teamwork/httperr/stack_test.go:37\n" +
			"github.com/teamwork/httperr.TestStackTraceFormat\n" +
			"\t.+/github.com/teamwork/httperr/stack_test.go:56",
	}, {
		stackTrace()[:2],
		"%#v",
		`\[\]errors.Frame{stack_test.go:37, stack_test.go:64}`,
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.StackTrace, tt.format, tt.want)
	}
}
