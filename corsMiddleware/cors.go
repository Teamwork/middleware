package corsMiddleware // import "github.com/teamwork/middleware/corsMiddleware"

import (
	"net/http"
	"strconv"
	"strings"
)

// HTTP methods
const (
	CONNECT = "CONNECT"
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
	TRACE   = "TRACE"
)

// Constants for header types
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

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

type (
	// Config defines the config for CORS middleware.
	Config struct {
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
)

var (
	// DefaultConfig is the default CORS middleware config.
	DefaultConfig = Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{GET, HEAD, PUT, POST, DELETE},
	}
)

// Add returns a Cross-Origin Resource Sharing (CORS) middleware.
// See https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS
func Add(f http.HandlerFunc) http.HandlerFunc {
	return WithConfig(DefaultConfig)(f)
}

// WithConfig returns a CORS middleware from config.
// See `CORS()`.
func WithConfig(config Config) func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Defaults
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

			origin := r.Header.Get(HeaderOrigin)

			// Check allowed origins
			allowedOrigin := ""
			for _, o := range config.AllowOrigins {
				if o == "*" || o == origin {
					allowedOrigin = o
					break
				}
			}

			// Simple request
			if r.Method != OPTIONS {
				w.Header().Add(HeaderVary, HeaderOrigin)
				if origin == "" || allowedOrigin == "" {
					f(w, r)
					return
				}
				w.Header().Set(HeaderAccessControlAllowOrigin, allowedOrigin)
				if config.AllowCredentials {
					w.Header().Set(HeaderAccessControlAllowCredentials, "true")
				}
				if exposeHeaders != "" {
					w.Header().Set(HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				f(w, r)
				return
			}

			// Preflight request
			w.Header().Add(HeaderVary, HeaderOrigin)
			w.Header().Add(HeaderVary, HeaderAccessControlRequestMethod)
			w.Header().Add(HeaderVary, HeaderAccessControlRequestHeaders)
			if origin == "" || allowedOrigin == "" {
				f(w, r)
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
			return
		}
	}
}
