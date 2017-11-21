package apiversionMiddleware

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

func TestWithConfig(t *testing.T) {
	cases := []struct {
		in         *http.Request
		conf       Config
		wantHeader string
	}{
		{
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{},
			},
			Config{
				AWSRegion: "a",
				Env:       "b",
				Version:   "c",
			},
			"region: 'a' env: 'b' version: 'c'",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			rr := test.HTTP(t, tc.in, WithConfig(tc.conf)(handle{}).ServeHTTP)

			v := rr.Header().Get("API-Version")

			if v != tc.wantHeader {
				t.Errorf("want API-Version %v, got %v", tc.wantHeader, v)
			}
		})
	}
}
