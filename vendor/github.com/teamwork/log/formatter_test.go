package log

import (
	goerr "errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/teamwork/test/diff"
)

var now time.Time

func init() {
	now, _ = time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
}

type formatTest struct {
	Name     string
	Entry    *logrus.Entry
	Expected string
}

func withPid(format string) string {
	return fmt.Sprintf(format, os.Getpid())
}

func TestColorFormatter(t *testing.T) {
	var formatter = &twFormatter{
		ForceColors: true,
	}
	tests := []formatTest{
		formatTest{
			Name: "simple entry with colored output",
			Entry: &logrus.Entry{
				Data:    logrus.Fields{},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid("Jan 02 15:04:05 [\x1b[57m%d\x1b[0m] \x1b[34mINFO\x1b[0m: Test log message\n"),
		},
	}
	runTests(t, formatter, tests)
}

func TestFormatter(t *testing.T) {
	var formatter = &twFormatter{
		DisableColors: true,
	}
	tests := []formatTest{
		formatTest{
			Name: "simple entry",
			Entry: &logrus.Entry{
				Data:    logrus.Fields{},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid("Jan 02 15:04:05 [%d] INFO: Test log message\n"),
		},
		formatTest{
			Name: "entry with module",
			Entry: &logrus.Entry{
				Data: logrus.Fields{
					fieldModule: "foo",
				},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid("Jan 02 15:04:05 [%d] INFO: (foo) Test log message\n"),
		},
		formatTest{
			Name: "entry with extra data",
			Entry: &logrus.Entry{
				Data: logrus.Fields{
					fieldModule: "foo",
					"animal":    "cow",
					"legs":      4,
				},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid(`Jan 02 15:04:05 [%d] INFO: (foo) Test log message
	animal="cow" legs=4
`),
		},
		formatTest{
			Name: "entry with simple error",
			Entry: &logrus.Entry{
				Data: logrus.Fields{
					fieldError: goerr.New("simple error"),
					"duck":     "quack",
				},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid(`Jan 02 15:04:05 [%d] INFO: Test log message
	duck="quack"
	Error: simple error
`),
		},
		formatTest{
			Name: "entry with caused error",
			Entry: &logrus.Entry{
				Data: logrus.Fields{
					fieldError: errors.Wrap(goerr.New("simple error"), "secondary error"),
					"duck":     "quack",
				},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid(`Jan 02 15:04:05 [%d] INFO: Test log message
	duck="quack"
	Error: secondary error: simple error
	Cause: simple error
	Stacktrace:
		github.com/teamwork/log.TestFormatter
			.../teamwork/log/formatter_test.go:999
		testing.tRunner
			.../src/testing/testing.go:999
		runtime.goexit
			.../src/runtime/asm_amd64.s:999
`),
		},
		formatTest{
			Name: "error_message without error",
			Entry: &logrus.Entry{
				Data: logrus.Fields{
					fieldErrorMessage: "orphaned error message",
					"duck":            "quack",
				},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid(`Jan 02 15:04:05 [%d] INFO: Test log message
	duck="quack" error_message="orphaned error message"
`),
		},
		formatTest{
			Name: "nil error",
			Entry: &logrus.Entry{
				Data: logrus.Fields{
					fieldError: nil,
				},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid(`Jan 02 15:04:05 [%d] INFO: Test log message
	error=<nil>
`),
		},
		formatTest{
			Name: "http request",
			Entry: &logrus.Entry{
				Data: logrus.Fields{
					fieldHTTPRequest: func() *http.Request {
						req, _ := http.NewRequest("GET", "http://example.com", nil)
						req.Header = http.Header(map[string][]string{
							"Content-Type": []string{"foo/bar"},
							"X-Shizzle":    []string{"nizzle", "wizzle"},
						})
						return req
					}(),
				},
				Time:    now,
				Level:   logrus.InfoLevel,
				Message: "Test log message",
			},
			Expected: withPid(`Jan 02 15:04:05 [%d] INFO: Test log message
	HTTP/1.1 Request: GET http://example.com
		Content-Type: foo/bar
		X-Shizzle: nizzle
		X-Shizzle: wizzle
`),
		},
	}
	runTests(t, formatter, tests)
}

type nilErrorWrapper struct {
	err error
}

func (ne *nilErrorWrapper) Error() string {
	if ne.err == nil {
		return ""
	}
	return ne.err.Error()
}

func (ne *nilErrorWrapper) Cause() error {
	return ne.err
}

var pathRE = regexp.MustCompile("^\t\t\t/.*\\.(go|s):\\d")

func runTests(t *testing.T, formatter logrus.Formatter, tests []formatTest) {
	for _, test := range tests {
		b, err := formatter.Format(test.Entry)
		if err != nil {
			t.Errorf("Error formatting test %s: %s", test.Name, err)
		}
		if diff := diff.TextDiff(test.Expected, sanitizeLog(string(b))); diff != "" {
			t.Errorf("Results differ:\n%s\n", diff)
		}
	}
}

func sanitizeLog(log string) string {
	var out []string
	for _, line := range strings.Split(log, "\n") {
		if pathRE.MatchString(line) {
			parts := strings.Split(line, "/")
			if len(parts) > 4 {
				parts = append([]string{"\t\t\t..."}, parts[len(parts)-3:len(parts)]...)
				line = strings.TrimRight(strings.Join(parts, "/"), "0123456789") + "999"
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
