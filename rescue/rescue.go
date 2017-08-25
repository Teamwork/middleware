// Package rescue is a middleware to recover() and log panic()s.
//
// It will also return an appropriate response to the client (HTML, JSON, or
// text).
package rescue // import "github.com/teamwork/middlware/rescue"

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/kr/pretty"
	"github.com/labstack/echo"
	"github.com/spf13/viper"

	"github.com/teamwork/log"
)

// Handle recovers and logs panics in any of the lower middleware or HTTP
// handlers.
//
// The extraFields callback can be used to add extra fields to the log (such as
// perhaps a installation ID or user ID from the session).
func Handle(extraFields func(echo.Context, *log.Entry) *log.Entry) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			l := log.Module("panic handler")
			defer func() {
				if r := recover(); r != nil {
					var err error
					switch r := r.(type) {
					case error:
						err = r
					case map[string]interface{}:
						err, _ = r["error"].(error)
					default:
						err = pretty.Errorf("%v", r)
					}

					if extraFields != nil {
						l = extraFields(c, l)
					}

					// Report to Sentry.
					l.Err(err)

					switch {

					// Show panic in browser on dev.
					case viper.GetBool("dev.enabled"):
						if c.Request().Header().Get("X-Requested-With") == "XMLHttpRequest" {
							c.JSON(http.StatusInternalServerError, err.Error())
							return
						}
						c.HTML(http.StatusInternalServerError,
							fmt.Sprintf("<h2>%v</h2><pre>%s</pre>",
								err, debug.Stack()))

					// JSON response for AJAX.
					case c.Request().Header().Get("X-Requested-With") == "XMLHttpRequest":
						c.JSON(http.StatusInternalServerError, map[string]interface{}{
							"message": "Sorry, the server ran into a problem processing this request.",
						})

					// Well, just fall back to text..
					default:
						c.String(http.StatusInternalServerError,
							"Sorry, the server ran into a problem processing this request.")
					}
				}
			}()
			return next(c)
		}
	}
}
