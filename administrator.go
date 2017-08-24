package middleware

import (
	"net/http"

	"github.com/labstack/echo"
)

// AdministratorLockdown ensures that the admin can only be accessed locally or
// on digitalcrew accounts
func AdministratorLockdown() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			host := c.Request().Host()
			if host == "digitalcrew.teamwork.com" ||
				host == "digitalcreweu.eu.teamwork.com" ||
				host == "sunbeam.teamwork.dev" {
				return next(c)
			}

			return c.HTML(http.StatusUnauthorized, "not authorized")
		}
	}
}
