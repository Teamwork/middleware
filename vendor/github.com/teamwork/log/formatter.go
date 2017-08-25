package log

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// TimeFormat to use for log messages.
const TimeFormat = "Jan 02 15:04:05"

type twFormatter struct {
	ForceColors   bool
	DisableColors bool
}

// borrowed from https://github.com/sirupsen/logrus/blob/master/text_formatter.go#L72-L79
func checkIfTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return terminal.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}

func (f *twFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	useColors := !f.DisableColors && (f.ForceColors || checkIfTerminal(os.Stdout))

	if err := printMessage(b, entry, useColors); err != nil {
		return nil, err
	}

	errBytes := extractError(entry.Data, useColors)
	reqBytes := extractRequest(entry.Data, useColors)

	if err := printData(b, entry.Data, useColors); err != nil {
		return nil, err
	}

	if _, err := b.Write(errBytes); err != nil {
		return nil, err
	}
	if _, err := b.Write(reqBytes); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

const (
	reset      = 0
	bold       = 1
	underscore = 4
	blink      = 5
	reverse    = 7
	concealed  = 8

	fgBlack   = 30
	fgRed     = 31
	fgGreen   = 32
	fgYellow  = 33
	fgBlue    = 34
	fgMagenta = 35
	fgCyan    = 36
	fgWhite   = 37

	bgBlack   = 40
	bgRed     = 41
	bgGreen   = 42
	bgYellow  = 43
	bgBlue    = 44
	bgMagenta = 45
	bgCyan    = 46
	bgWhite   = 57
)

var colors = map[logrus.Level]uint8{
	logrus.DebugLevel: fgWhite,
	logrus.ErrorLevel: fgRed,  // red
	logrus.InfoLevel:  fgBlue, // blue
}

func colorize(text string, color uint8, colored bool) string {
	if !colored {
		return text
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, text)
}

func printMessage(b *bytes.Buffer, entry *logrus.Entry, color bool) error {
	levelText := colorize(
		strings.ToUpper(entry.Level.String())[0:4],
		colors[entry.Level],
		color,
	)
	module, ok := entry.Data[fieldModule].(string)
	if ok {
		delete(entry.Data, fieldModule) // So it isn't repeated
		module = " (" + module + ")"
	}
	pid := colorize(
		fmt.Sprintf("%d", os.Getpid()),
		bgWhite,
		color,
	)
	_, err := fmt.Fprintf(b, "%s [%s] %s:%s %s\n",
		entry.Time.Format(TimeFormat), pid, levelText, module,
		strings.TrimSpace(entry.Message))
	return err
}

func printData(b *bytes.Buffer, data map[string]interface{}, color bool) error {
	if len(data) == 0 {
		return nil
	}
	if err := b.WriteByte('\t'); err != nil {
		return err
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	//    unhandled := make(map[string]interface{})

	pairs := make([][]byte, 0, len(data))
	for _, k := range keys {
		v := data[k]
		var formatted string
		switch v.(type) {
		case string:
			formatted = fmt.Sprintf("%s=%q", k, v)
		case url.URL, *url.URL:
			formatted = fmt.Sprintf("%s=%s", k, v)
		default:
			formatted = fmt.Sprintf("%s=%#v", k, v)
		}
		pairs = append(pairs, []byte(formatted))
	}
	if _, err := b.Write(bytes.Join(pairs, []byte(" "))); err != nil {
		return err
	}

	return b.WriteByte('\n')
}

func extractError(data map[string]interface{}, color bool) []byte {
	err, ok := data[fieldError].(error)
	if !ok {
		return nil
	}
	delete(data, fieldError)
	if err == nil {
		return nil
	}
	// The error message field is just for sentry; it's a duplicate
	// of err.Error(), which we handle separately here.
	delete(data, fieldErrorMessage)
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "\tError: %s\n", err.Error())
	if _, ok := err.(causer); ok {
		cause := errors.Cause(err)
		if cause != nil &&
			// No need to repeat an identical message
			cause.Error() != err.Error() {
			fmt.Fprintf(b, "\tCause: %s\n", cause.Error())
		}
	}
	if tracer := earliestStackTracer(err); tracer != nil {
		fmt.Fprint(b, "\tStacktrace:\n")
		for _, line := range strings.Split(fmt.Sprintf("%+v", tracer.StackTrace()), "\n") {
			if line != "" {
				fmt.Fprintf(b, "\t\t%s\n", line)
			}
		}
	}
	return b.Bytes()
}

func extractRequest(data map[string]interface{}, color bool) []byte {
	req, ok := data[fieldHTTPRequest].(*http.Request)
	if !ok {
		return nil
	}
	delete(data, fieldHTTPRequest)
	b := &bytes.Buffer{}
	method := colorize(req.Method, fgCyan, color)
	url := colorize(req.URL.String(), bold, color)
	fmt.Fprintf(b, "\t%s Request: %s %s\n", req.Proto, method, url)
	var headers []string
	for h := range req.Header {
		headers = append(headers, h)
	}
	sort.Strings(headers)
	for _, k := range headers {
		for _, v := range req.Header[k] {
			header := colorize(k, fgCyan, color)
			value := colorize(v, bold, color)
			fmt.Fprintf(b, "\t\t%s: %s\n", header, value)
		}
	}
	return b.Bytes()
}
