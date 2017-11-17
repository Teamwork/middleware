package httperrMiddleware // import "github.com/teamwork/middleware/httperrMiddleware"

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestErrType(t *testing.T) {
	type typeTest struct {
		name     string
		req      *http.Request
		res      http.ResponseWriter
		expected string
	}
	tests := []typeTest{
		{
			name:     "NoHeaders",
			req:      httptest.NewRequest("GET", "/", nil),
			res:      httptest.NewRecorder(),
			expected: "application/json",
		},
		{
			name: "HTMLResponse",
			req:  httptest.NewRequest("GET", "/", nil),
			res: func() http.ResponseWriter {
				res := httptest.NewRecorder()
				res.Header().Set("Content-Type", "text/html")
				return res
			}(),
			expected: "text/html",
		},
		{
			name: "AcceptHTML",
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept", "text/html")
				return req
			}(),
			res:      httptest.NewRecorder(),
			expected: "text/html",
		},
		{
			name: "AcceptGIF",
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept", "image/gif")
				return req
			}(),
			res:      httptest.NewRecorder(),
			expected: "application/json",
		},
		{
			name: "AcceptAnyText",
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept", "text/*")
				return req
			}(),
			res:      httptest.NewRecorder(),
			expected: "text/plain",
		},
		{
			name: "AcceptGIFthenText",
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept", "image/gif, text/plain")
				return req
			}(),
			res:      httptest.NewRecorder(),
			expected: "text/plain",
		},
		{
			name: "AcceptHTMLorText",
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept", "text/html; q=0.5, text/plain; q=0.8")
				return req
			}(),
			res:      httptest.NewRecorder(),
			expected: "text/plain",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := errType(test.req, test.res)
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}

type warning struct{}

func (w warning) Error() string {
	return "oh noes"
}
func (w warning) IsWarning() bool {
	return true
}

func TestHTTPErrorHandler(t *testing.T) {
	type errTest struct {
		name     string
		req      *http.Request
		res      *httptest.ResponseRecorder
		err      error
		status   int
		expected string
	}
	tests := []errTest{
		{
			name:     "NoCode",
			req:      httptest.NewRequest("GET", "/", nil),
			err:      errors.New("standard error"),
			status:   500,
			expected: `{"errors":["standard error"],"status":"error"}`,
		},
		{
			name:     "HEAD",
			req:      httptest.NewRequest("HEAD", "/", nil),
			err:      errors.New("standard error"),
			status:   500,
			expected: "",
		},
		{
			name:     "warning",
			req:      httptest.NewRequest("GET", "/", nil),
			err:      errors.New("a warning"),
			status:   500,
			expected: `{"errors":["a warning"],"status":"error"}`,
		},
		{
			name: "TextResponse",
			req:  httptest.NewRequest("GET", "/", nil),
			res: func() *httptest.ResponseRecorder {
				res := httptest.NewRecorder()
				res.Header().Set("Content-Type", "text/plain")
				return res
			}(),
			err:      errors.New("standard error"),
			status:   500,
			expected: "standard error",
		},
		{
			name:     "Warning",
			req:      httptest.NewRequest("GET", "/", nil),
			err:      warning{},
			status:   500,
			expected: `{"errors":["oh noes"],"status":"error"}`,
		},
		{
			name:     "warning",
			req:      httptest.NewRequest("GET", "/", nil),
			err:      context.DeadlineExceeded,
			status:   504,
			expected: `{"errors":["Sorry, something is taking too long. Please try again"],"status":"error"}`,
		},
		{
			name:     "warning",
			req:      httptest.NewRequest("GET", "/", nil),
			err:      &echo.HTTPError{Code: 502, Message: "woot"},
			status:   502,
			expected: `{"errors":["woot"],"status":"error"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := test.res
			if w == nil {
				w = httptest.NewRecorder()
			}
			c := echo.New().NewContext(test.req, w)
			ErrorHandler(test.err, c)
			res := w.Result()
			if res.StatusCode != test.status {
				t.Errorf("Unexpected status: %d %s", res.StatusCode, res.Status)
			}
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			err = res.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			if test.expected != string(resBody) {
				t.Errorf("\ngot:  %v\nwant: %v\n", string(resBody), test.expected)
			}
		})
	}
}
