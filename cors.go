package middleware // import "github.com/teamwork/middleware"

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

type (
	// CORSConfig defines the config for CORS middleware.
	CORSConfig struct {
		// AllowOrigin defines a list of origins that may access the resource.
		// Optional, with default value as []string{"*"}.
		AllowOrigins []string

		// AllowMethods defines a list methods allowed when accessing the resource.
		// This is used in response to a preflight request.
		// Optional, with default value as `DefaultCORSConfig.AllowMethods`.
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
)

var (
	// DefaultCORSConfig is the default CORS middleware config.
	DefaultCORSConfig = CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.POST, echo.DELETE},
	}
)

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware.
// See https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS
func CORS(f http.HandlerFunc) http.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig)(f)
}

// CORSWithConfig returns a CORS middleware from config.
// See `CORS()`.
func CORSWithConfig(config CORSConfig) func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Defaults
			if len(config.AllowOrigins) == 0 {
				config.AllowOrigins = DefaultCORSConfig.AllowOrigins
			}
			if len(config.AllowMethods) == 0 {
				config.AllowMethods = DefaultCORSConfig.AllowMethods
			}
			allowMethods := strings.Join(config.AllowMethods, ",")
			allowHeaders := strings.Join(config.AllowHeaders, ",")
			exposeHeaders := strings.Join(config.ExposeHeaders, ",")
			maxAge := strconv.Itoa(config.MaxAge)

			origin := r.Header.Get(echo.HeaderOrigin)

			// Check allowed origins
			allowedOrigin := ""
			for _, o := range config.AllowOrigins {
				if o == "*" || o == origin {
					allowedOrigin = o
					break
				}
			}

			// Simple request
			if r.Method != echo.OPTIONS {
				w.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
				if origin == "" || allowedOrigin == "" {
					f(w, r)
					return
				}
				w.Header().Set(echo.HeaderAccessControlAllowOrigin, allowedOrigin)
				if config.AllowCredentials {
					w.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
				}
				if exposeHeaders != "" {
					w.Header().Set(echo.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				f(w, r)
				return
			}

			// Preflight request
			w.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
			w.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
			w.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
			if origin == "" || allowedOrigin == "" {
				f(w, r)
				return
			}
			w.Header().Set(echo.HeaderAccessControlAllowOrigin, allowedOrigin)
			w.Header().Set(echo.HeaderAccessControlAllowMethods, allowMethods)
			if config.AllowCredentials {
				w.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
			}
			if allowHeaders != "" {
				w.Header().Set(echo.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				h := r.Header.Get(echo.HeaderAccessControlRequestHeaders)
				if h != "" {
					w.Header().Set(echo.HeaderAccessControlAllowHeaders, h)
				}
			}
			if config.MaxAge > 0 {
				w.Header().Set(echo.HeaderAccessControlMaxAge, maxAge)
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
}
