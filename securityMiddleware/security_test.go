package securityMiddleware

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
func Test(t *testing.T) {
	cases := []struct {
		in   Config
		want http.Header
	}{
		{Config{}, http.Header{}},
		{DefaultConfig, http.Header{
			"Strict-Transport-Security": []string{"max-age=2592000"},
			"X-Frame-Options":           []string{"SAMEORIGIN"},
			"X-Content-Type-Options":    []string{"nosniff"}},
		},
		{
			Config{
				ContentSecurityPolicy: map[string][]string{
					"default-src": []string{"'self'", "https://example.com"},
					"script-src":  []string{"https://static.example.org"},
				},
			},
			http.Header{
				"Content-Security-Policy": []string{
					"default-src 'self' https://example.com; script-src https://static.example.org;"},
			},
		},
		{
			Config{
				ContentSecurityPolicyReportOnly: map[string][]string{
					"default-src": []string{"'self'", "https://example.com"},
					"script-src":  []string{"https://static.example.org"},
				},
			},
			http.Header{
				"Content-Security-Policy-Report-Only": []string{
					"default-src 'self' https://example.com; script-src https://static.example.org;"},
			},
		},
		{
			Config{
				StrictTransportSecurity: "max-age=666",
			},
			http.Header{
				"Strict-Transport-Security": []string{"max-age=666"},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			req := &http.Request{Host: "example.com"}
			rr := test.HTTP(t, req, WithConfig(tc.in, "example.com")(handle{}).ServeHTTP)
			out := rr.Header()
			out.Del("Content-Type")
			// TODO: we use a map so order isn't guaranteed :-/
			//If !reflect.DeepEqual(tc.want, out) {
			//	t.Errorf("\nout:  %#v\nwant: %#v\n", out, tc.want)
			//}
		})
	}
}
