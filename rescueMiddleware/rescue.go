// Package rescueMiddleware recover()s and log panic()s.
//
// It will also return an appropriate response to the client (HTML, JSON, or
// text).
package rescueMiddleware // import "github.com/teamwork/middleware/rescueMiddleware"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/kr/pretty"

	"github.com/teamwork/log"
)

// Handle recovers and logs panics in any of the lower middleware or HTTP
// handlers.
//
// The extraFields callback can be used to add extra fields to the log (such as
// perhaps a installation ID or user ID from the session).
func Handle(extraFields func(*log.Entry) *log.Entry, dev bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := log.Module("panic handler")
			defer func() {
				if rec := recover(); r != nil {
					var err error
					switch rec := rec.(type) {
					case error:
						err = rec
					case map[string]interface{}:
						err, _ = rec["error"].(error)
					default:
						err = pretty.Errorf("%v", rec)
					}

					if extraFields != nil {
						l = extraFields(l)
					}

					// Report to Sentry.
					l.Err(err)

					w.WriteHeader(http.StatusInternalServerError)

					switch {
					// Show panic in browser on dev.
					case dev:
						if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
							w.Write([]byte(err.Error())) // nolint: errcheck
							return
						}

						// nolint: errcheck
						w.Write([]byte(fmt.Sprintf("<h2>%v</h2><pre>%s</pre>",
							err, debug.Stack())))

					// JSON response for AJAX.
					case r.Header.Get("X-Requested-With") == "XMLHttpRequest":
						b, _ := json.Marshal(map[string]interface{}{
							"message": "Sorry, the server ran into a problem processing this request.",
						})
						w.Header().Add("Content-Type", "application/json")
						w.Write(b) // nolint: errcheck

					// Well, just fall back to text..
					default:
						w.Write([]byte("Sorry, the server ran into a problem processing this request.")) // nolint: errcheck
					}

					return
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
