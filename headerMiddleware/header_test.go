package headerMiddleware

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

func TestSet(t *testing.T) {
	cases := []struct {
		in         *http.Request
		conf       http.Header
		wantHeader string
	}{
		{
			&http.Request{Header: http.Header{}},
			http.Header{"API-Version": []string{"w00t"}},
			"w00t",
		},
		{
			&http.Request{Header: http.Header{}},
			http.Header{"API-Version": []string{"w00t", "second"}},
			"second",
		},
		{
			&http.Request{Header: http.Header{"API-Version": []string{"existing"}}},
			http.Header{"API-Version": []string{"w00t"}},
			"w00t",
		},
		{
			&http.Request{Header: http.Header{}},
			http.Header{"api-version": []string{"w00t"}},
			"w00t",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			rr := test.HTTP(t, tc.in, Set(tc.conf)(handle{}).ServeHTTP)

			v := rr.Header().Get("API-Version")
			if v != tc.wantHeader {
				t.Errorf("want API-Version %v, got %v", tc.wantHeader, v)
			}
		})
	}
}
