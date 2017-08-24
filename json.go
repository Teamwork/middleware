package middleware

import (
	"mime"
	"net/http"

	"github.com/labstack/echo"
	"github.com/teamwork/desk/modules/httperr"
)

// RequireJSON returns http.StatusUnsupportedMediaType if the request is not
// of type application/json.
func RequireJSON() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if ct, _, _ := mime.ParseMediaType(c.Request().Header().Get("Content-Type")); ct != "application/json" {
				return httperr.NewWithCode(http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			}

			return next(c)
		}
	}
}
