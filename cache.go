package middleware

import "github.com/labstack/echo"

// NoCache sets the Cache-Control header to "no-cache" to prevent browsing caching.
func NoCache() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Response().Header().Set("Cache-Control", "no-cache")
			return next(c)
		}
	}
}
