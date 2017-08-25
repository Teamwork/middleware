package log

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

const maxStackFrames = 25

var myPath string

func init() {
	_, file, _, _ := runtime.Caller(1)
	myPath = filepath.Dir(file)
}

type stackFrame struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

func callStack(count int) []stackFrame {
	pc := make([]uintptr, maxStackFrames)
	runtime.Callers(1, pc)
	return stackFrames(pc, count)
}

func errorStack(err stackTracer) []stackFrame {
	stacktrace := err.StackTrace()
	pc := make([]uintptr, len(stacktrace))
	for i := range stacktrace {
		pc[i] = uintptr(stacktrace[i])
	}
	return stackFrames(pc, -1)
}

type withStack struct {
	err   error
	stack errors.StackTrace
}

func (w *withStack) Cause() error {
	return w.err
}

func (w *withStack) StackTrace() errors.StackTrace {
	return w.stack
}

func (w *withStack) Error() string {
	if w.err == nil {
		return ""
	}
	return w.err.Error()
}

func addStackTrace(err error) error {
	if _, ok := err.(stackTracer); ok {
		return err
	}
	pc := make([]uintptr, maxStackFrames)
	count := runtime.Callers(1, pc)
	var i int
	for ; i < count; i++ {
		fn := runtime.FuncForPC(pc[i])
		file, _ := fn.FileLine(pc[i])
		if !strings.HasPrefix(file, myPath) || strings.HasSuffix(file, "_test.go") {
			break
		}
	}
	stack := make([]errors.Frame, count-i)
	for j, ptr := range pc[i:count] {
		stack[j] = errors.Frame(ptr)
	}
	return &withStack{
		err:   err,
		stack: stack,
	}
}

func stackFrames(pc []uintptr, count int) []stackFrame {
	stack := make([]stackFrame, 0, maxStackFrames)
	frames := runtime.CallersFrames(pc)
	for {
		frame, more := frames.Next()
		if frame.PC > 0 {
			if !strings.HasSuffix(frame.File, "_test.go") && strings.HasPrefix(frame.File, myPath) {
				// Skip frames from within this package, except for tests
				continue
			}
			f := stackFrame{
				Function: frame.Function,
				File:     frame.File,
				Line:     frame.Line,
			}
			stack = append(stack, f)
			if len(stack) == count {
				break
			}
		}
		if !more {
			break
		}
	}
	return stack
}

func callerFields() Fields {
	frames := callStack(10)
	for _, frame := range frames {
		if !strings.Contains(frame.File, "_test/_obj_test/") { // Skip testing frames
			return Fields{
				"file":     frame.File,
				"line":     frame.Line,
				"function": frame.Function,
			}
		}
	}
	return Fields{
		"file":     frames[0].File,
		"line":     frames[0].Line,
		"function": frames[0].Function,
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

func errorStackTraceFields(err error) Fields {
	fields := make(map[string]interface{})
	var stackTrace []stackFrame
	stackTrace = errorStackTrace(err)
	fields["stackTrace"] = stackToFields(stackTrace)
	return fields
}

func earliestStackTracer(err error) stackTracer {
	var tracer stackTracer
	for err != nil {
		if st, ok := err.(stackTracer); ok {
			tracer = st
		}
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return tracer
}

func errorStackTrace(err error) []stackFrame {
	if tracer := earliestStackTracer(err); tracer != nil {
		return errorStack(tracer)
	}
	return callStack(-1)
}

func stackToFields(stack []stackFrame) []map[string]interface{} {
	frames := make([]map[string]interface{}, 0, len(stack))
	for _, frame := range stack {
		frames = append(frames, map[string]interface{}{
			"file":     frame.File,
			"line":     frame.Line,
			"function": frame.Function,
		})
	}
	return frames
}
