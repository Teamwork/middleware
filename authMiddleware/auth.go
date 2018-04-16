// Package authMiddleware provides HTTP authentication.
package authMiddleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"
)

// Options for auth.
type Options struct {
	Username string
	Password string
	Realm    string
}

// Auth adds HTTP Basic authentication.
func Auth(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		username := []byte(opts.Username)
		password := []byte(opts.Password)
		realm := `Basic realm="Restricted"`
		if opts.Realm != "" {
			realm = fmt.Sprintf(`Basic realm="%v"`, opts.Realm)
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok ||
				subtle.ConstantTimeCompare([]byte(user), username) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), password) != 1 {

				w.Header().Set("WWW-Authenticate", realm)
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Unauthorised.\n"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
