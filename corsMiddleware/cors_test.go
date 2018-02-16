package corsMiddleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/teamwork/test"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))
}

func TestDo(t *testing.T) {
	cases := []struct {
		req  http.Request
		opts Config
	}{
		{
			http.Request{Method: "GET"},
			DefaultConfig,
		},
		{
			http.Request{Method: "POST"},
			DefaultConfig,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			rr := test.HTTP(t, &tc.req, Add(tc.opts)(handle{}).ServeHTTP)

			if rr.Code != 200 {
				t.Fatalf("code != 200: %v", rr.Code)
			}
			body := rr.Body.String()
			if body != "handler" {
				t.Fatalf("body != handler: %v", body)
			}
		})
	}
}
