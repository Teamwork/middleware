package apiversionMiddleware // import "github.com/teamwork/middleware/apiversionMiddleware"

import (
	"fmt"
	"net/http"
)

// Config are server information fields
type Config struct {
	AWSRegion string
	Env       string
	Version   string
}

// WithConfig adds region, git commit & beta/prod to the server response
func WithConfig(c Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("API-Version", fmt.Sprintf("region: '%s' env: '%s' version: '%s'", c.AWSRegion, c.Env, c.Version))
			next.ServeHTTP(w, r)
		})
	}
}
