// Package security adds various security-related headers.
package security // import "github.com/teamwork/middleware/security"

import (
	"fmt"
	"net/http"
	"strings"
)

// Config defines the config for Security middleware.
type Config struct {
	// XFrameOptions controls where this site can be displayed in a frame.
	//
	// DENY
	//     The page cannot be displayed in a frame, regardless of the site
	//     attempting to do so.
	// SAMEORIGIN
	//     The page can only be displayed in a frame on the same origin as the
	//     page itself.
	// ALLOW-FROM uri
	//     The page can only be displayed in a frame on the specified origin.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
	XFrameOptions string

	// StrictTransportSecurity makes sure that browsers only communicate over
	// https. It will only be set if the host matches the root domain.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
	StrictTransportSecurity string

	// XContentTypeOptions makes sure that browsers don't auto-guess the
	// Content-Type, preventing certain attacks.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
	XContentTypeOptions bool

	// ContentSecurityPolicy controls which JS and CSS resources can be run,
	// preventing XSS attacks
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
	ContentSecurityPolicy map[string][]string

	// ContentSecurityPolicyReportOnly is like the CSP header, but only reports
	// violations and doesn't block anything, which is useful for testing new
	// policies.
	ContentSecurityPolicyReportOnly map[string][]string
}

// DefaultConfig is the default Security middleware config.
var DefaultConfig = Config{
	XFrameOptions:           "SAMEORIGIN",      // Allow displaying in frame only on same domain
	StrictTransportSecurity: "max-age=2592000", // Only allow http for the next 30 days
	XContentTypeOptions:     true,              // Block CSS/JS files without correct Content-Type
}

// WithConfig returns a Security middleware from config.
func WithConfig(config Config, rootDomain string) func(http.Handler) http.Handler {
	// Avoid looping over the map on every request.
	csp := ""
	for k, v := range config.ContentSecurityPolicy {
		csp += fmt.Sprintf("%v %v; ", k, strings.Join(v, " "))
	}
	csp = strings.TrimRight(csp, " ")

	cspReport := ""
	for k, v := range config.ContentSecurityPolicyReportOnly {
		cspReport += fmt.Sprintf("%v %v; ", k, strings.Join(v, " "))
	}
	cspReport = strings.TrimRight(cspReport, " ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if config.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.XFrameOptions)
			}
			if csp != "" {
				w.Header().Set("Content-Security-Policy", csp)
			}
			if cspReport != "" {
				w.Header().Set("Content-Security-Policy-Report-Only", cspReport)
			}
			if config.StrictTransportSecurity != "" &&
				strings.HasSuffix(r.Host, rootDomain) {
				w.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
			}
			if config.XContentTypeOptions {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			next.ServeHTTP(w, r)
		})
	}
}
