// Package header adds HTTP headers.
package header // import "github.com/teamwork/middleware/header"

import (
	"net/http"
)

// Set header values, overwriting any previous value.
func Set(h http.Header) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, vals := range h {
				for _, v := range vals {
					w.Header().Set(k, v)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
