package authMiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))
}

func TestAuth(t *testing.T) {
	cases := []struct {
		username, password string
		expectedCode       int
		expectedBody       string
	}{
		{"", "", http.StatusUnauthorized, "Unauthorised.\n"},
		{"asd", "qweX", http.StatusUnauthorized, "Unauthorised.\n"},
		{"asdX", "qwe", http.StatusUnauthorized, "Unauthorised.\n"},
		{"asd", "", http.StatusUnauthorized, "Unauthorised.\n"},
		{"", "qwe", http.StatusUnauthorized, "Unauthorised.\n"},
		{"asd", "qwe", http.StatusOK, "handler"},
	}

	opts := Options{
		Username: "asd",
		Password: "qwe",
	}

	for _, tc := range cases {
		t.Run(tc.username+tc.password, func(t *testing.T) {
			rr := httptest.NewRecorder()

			handler := Auth(opts)(handle{}).ServeHTTP

			req, err := http.NewRequest("GET", "", nil)
			if err != nil {
				t.Fatalf("cannot make request: %v", err)
			}
			if tc.username != "" || tc.password != "" {
				req.SetBasicAuth(tc.username, tc.password)
			}

			handler(rr, req)
			if rr.Code != tc.expectedCode {
				t.Errorf("wrong status code: expected %v, got %v",
					tc.expectedCode, rr.Code)
			}

			body := rr.Body.String()
			if body != tc.expectedBody {
				t.Errorf("wrong body; expected %#v, got %#v",
					tc.expectedBody, body)
			}
		})
	}
}
