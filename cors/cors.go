// Package cors implements cross origin resource sharing (CORS) headers.
package cors // import "github.com/teamwork/middleware/cors"

import (
	"net/http"
	"strconv"
	"strings"
)

// Constants for header types
const (
	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

// Config for the middleware.
type Config struct {
	// AllowOrigin defines a list of origins that may access the resource.
	// Optional, with default value as []string{"*"}.
	AllowOrigins []string

	// AllowMethods defines a list methods allowed when accessing the resource.
	// This is used in response to a preflight request.
	// Optional, with default value as `DefaultConfig.AllowMethods`.
	AllowMethods []string

	// AllowHeaders defines a list of request headers that can be used when
	// making the actual request. This in response to a preflight request.
	// Optional, with default value as []string{}.
	AllowHeaders []string

	// AllowCredentials indicates whether or not the response to the request
	// can be exposed when the credentials flag is true. When used as part of
	// a response to a preflight request, this indicates whether or not the
	// actual request can be made using credentials.
	// Optional, with default value as false.
	AllowCredentials bool

	// ExposeHeaders defines a whitelist headers that clients are allowed to
	// access.
	// Optional, with default value as []string{}.
	ExposeHeaders []string

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	// Optional, with default value as 0.
	MaxAge int
}

// DefaultConfig is the default CORS middleware config.
var DefaultConfig = Config{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut,
		http.MethodPost, http.MethodDelete},
}

// Add the CORS middleware.
func Add(config Config) func(http.Handler) http.Handler {

	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultConfig.AllowOrigins
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultConfig.AllowMethods
	}

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			origin := r.Header.Get("Origin")

			// Check allowed origins
			allowedOrigin := ""
			for _, o := range config.AllowOrigins {
				if o == "*" || o == origin {
					allowedOrigin = o
					break
				}
			}

			// Simple request
			if r.Method != http.MethodOptions {
				w.Header().Add("Vary", "Origin")
				if origin == "" || allowedOrigin == "" {
					next.ServeHTTP(w, r)
					return
				}
				w.Header().Set(HeaderAccessControlAllowOrigin, allowedOrigin)
				if config.AllowCredentials {
					w.Header().Set(HeaderAccessControlAllowCredentials, "true")
				}
				if exposeHeaders != "" {
					w.Header().Set(HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				next.ServeHTTP(w, r)
				return
			}

			// Pre-flight request
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", HeaderAccessControlRequestMethod)
			w.Header().Add("Vary", HeaderAccessControlRequestHeaders)
			if origin == "" || allowedOrigin == "" {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set(HeaderAccessControlAllowOrigin, allowedOrigin)
			w.Header().Set(HeaderAccessControlAllowMethods, allowMethods)
			if config.AllowCredentials {
				w.Header().Set(HeaderAccessControlAllowCredentials, "true")
			}
			if allowHeaders != "" {
				w.Header().Set(HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				h := r.Header.Get(HeaderAccessControlRequestHeaders)
				if h != "" {
					w.Header().Set(HeaderAccessControlAllowHeaders, h)
				}
			}
			if config.MaxAge > 0 {
				w.Header().Set(HeaderAccessControlMaxAge, maxAge)
			}

			w.WriteHeader(http.StatusNoContent)
		})
	}
}
