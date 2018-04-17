package sanitize

import (
	"regexp"
	"sync"

	"github.com/labstack/echo"
)

var reURLExtensionsMatch = regexp.MustCompile(`(?i)\.(json|pdf|csv|zip)$`)
var extMU sync.Mutex

// StripFileExtensions strip .json from parameter names and values.
func StripFileExtensions() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// We only want to fix the last parameter, e.g.:
			// "/foo/:bar/:id.json"
			last := len(c.ParamNames()) - 1
			if last < 0 {
				return next(c)
			}

			if reURLExtensionsMatch.MatchString(c.Path()) {
				// Make sure that "/foo/42" does *NOT* work when the route is
				// "/foo/:id.json".
				if !reURLExtensionsMatch.MatchString(c.Request().URL().Path()) {
					return echo.ErrNotFound
				}

				extMU.Lock()
				if paramName := c.ParamNames()[last]; reURLExtensionsMatch.MatchString(paramName) {
					match := reURLExtensionsMatch.FindString(paramName)
					c.ParamNames()[last] = paramName[:len(paramName)-len(match)]
				}

				if paramValue := c.ParamValues()[last]; reURLExtensionsMatch.MatchString(paramValue) {
					match := reURLExtensionsMatch.FindString(paramValue)
					c.ParamValues()[last] = paramValue[:len(paramValue)-len(match)]
				}
				extMU.Unlock()
			}

			return next(c)
		}
	}
}
