package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

func TestRequireJSON(t *testing.T) {
	type rjTest struct {
		Name    string
		Request *http.Request
		Error   string
		IsJSON  bool
	}
	tests := []rjTest{
		{
			Name:    "NoContentType",
			Request: httptest.NewRequest("GET", "/foo", nil),
			Error:   "Content-Type must be application/json",
			IsJSON:  false,
		},
		{
			Name: "Image",
			Request: func() *http.Request {
				req := httptest.NewRequest("GET", "/foo", nil)
				req.Header.Set("Content-Type", "image/jpeg")
				return req
			}(),
			Error:  "Content-Type must be application/json",
			IsJSON: false,
		},
		{
			Name: "JSON",
			Request: func() *http.Request {
				req := httptest.NewRequest("GET", "/foo", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			IsJSON: true,
		},
		{
			Name: "JSON+charset",
			Request: func() *http.Request {
				req := httptest.NewRequest("GET", "/foo", nil)
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			}(),
			IsJSON: true,
		},
	}
	e := echo.New()
	for _, test := range tests {
		func(test rjTest) {
			t.Run(test.Name, func(t *testing.T) {
				w := httptest.NewRecorder()
				var isJSON bool
				handler := RequireJSON()(func(_ echo.Context) error {
					isJSON = true
					return nil
				})
				var msg string
				echoReq := standard.NewRequest(test.Request, nil)
				echoRes := standard.NewResponse(w, nil)
				if err := handler(e.NewContext(echoReq, echoRes)); err != nil {
					msg = err.Error()
				}
				if msg != test.Error {
					t.Errorf("Unexpected error: %s", msg)
				}

				if test.IsJSON != isJSON {
					t.Errorf("Expected %t, got %t", test.IsJSON, isJSON)
				}

			})
		}(test)
	}
}
