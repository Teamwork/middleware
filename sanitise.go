package middleware

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

// RemoveFilterStruct strips "filters." from query parameters.
//
// In revel Desk we accepted parameters like "filters.Sample" and "filters.*"
// because this is how Revel demarshaled the params. We're no longer really
// going to do this with Echo but there's no eloquent way to maintain
// compatibility.
// So our fix is to simply strip it from the path on the routes as required.
// Mainly all report routes and some other stat routes.
func RemoveFilterStruct() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request().(*standard.Request).Request
			req.URL.RawQuery = strings.Replace(req.URL.RawQuery, "filters.", "", -1)
			return next(c)
		}
	}
}

// DeindexFormFields converts indexed array params to unindexed arrays.
// types[0]=test becomes types[]=test.
func DeindexFormFields() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request().(*standard.Request).Request
			req.ParseForm()
			req.Form = replaceIndexedFormFields(req.Form)
			return next(c)
		}
	}
}

var reArrayIndexMatcher = regexp.MustCompile(`(.*)\[\d+\]`)

func replaceIndexedFormFields(form url.Values) url.Values {
	for key, values := range form {
		if reArrayIndexMatcher.MatchString(key) {
			matches := reArrayIndexMatcher.FindStringSubmatch(key)
			if len(matches) == 2 {
				newField := fmt.Sprintf("%s[]", matches[1])
				for _, value := range values {
					form.Add(newField, value)
				}

				form.Del(key)
			}
		}
	}

	return form
}

var reURLExtensionsMatch = regexp.MustCompile(`(?i)\.(json|pdf|csv|zip)$`)

var extMU sync.Mutex

// StripFileExtensions strips .json from parameter names and values.
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

// StripTrailingSlash strips trailing slashes from the path.
//
// We need to do this as Revel was also not sensitive about trailing slashes. It
// works by removing trailing slashes from the request URL before it is sent to
// the router.
func StripTrailingSlash() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL().Path()
			if strings.HasSuffix(path, "/") {
				c.Request().URL().SetPath(path[:len(path)-1])
			}
			return next(c)
		}
	}
}
