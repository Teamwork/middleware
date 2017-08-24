package reqlog

import (
	"fmt"
	"time"

	"github.com/labstack/echo"
	"github.com/teamwork/log"
)

// Log requests to stdout. This is mainly intended for dev-env.
func Log() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if err = next(c); err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			// Don't log health requests since there are so many!
			if req.URI() == "/health.json" {
				return nil
			}

			status := " "
			switch {
			case res.Status() >= 200 && res.Status() < 400:
				status = "\x1b[48;5;154m\x1b[38;5;0m%v\x1b[0m"
			case res.Status() >= 400 && res.Status() <= 499:
				status = "\x1b[1m\x1b[48;5;221m\x1b[38;5;0m%v\x1b[0m"
			case res.Status() >= 500 && res.Status() <= 599:
				status = "\x1b[1m\x1b[48;5;9m\x1b[38;5;15m%v\x1b[0m"
			}

			fmt.Printf("%v %v %v   %v%v\n",
				time.Now().Format(log.TimeFormat),
				fmt.Sprintf(status, res.Status()), req.Method(), req.Host(),
				req.URI())

			return nil
		}
	}
}
