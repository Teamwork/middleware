package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/teamwork/desk/metrics"
)

// Metrics adds this request to our metrics endpoint
// for collection via prometheus.
func Metrics() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			startTime := time.Now()
			err = next(c)
			method := c.Request().Method()
			status := strconv.Itoa(c.Response().Status())
			// Path may not be the most valuable URI component to track here.
			path := c.Path()
			metrics.OnRequestEnd(startTime, method, status, path)
			return err
		}
	}
}
