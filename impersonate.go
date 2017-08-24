package middleware

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/teamwork/desk/modules/httperr"
)

// BlockImpersonation is middleware used to block the impersonation feature of
// Projects from being used in Desk.
func BlockImpersonation() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			// If the TW-IMPERSONATE cookie is set to 1, we want to block the
			// person from being authorised in Desk. This is obviously easily
			// bypassed since it's a client-sided check, however it's the best
			// solution we can come up with for now without doing a lot of
			// changes to how the global sessions work.
			if cookie, err := c.Cookie("TW-IMPERSONATE"); err == nil {
				value, err := strconv.Atoi(cookie.Value())
				if err == nil && value != 0 {
					return httperr.NewWithCode(http.StatusNotFound, "Impersonation is not allowed in Desk")
				}
			}

			return next(c)
		}
	}
}
