package middleware

import (
	"github.com/labstack/echo"
	"github.com/spf13/viper"

	"github.com/teamwork/desk/models"
)

// SmartInboxesEnabled checks to see if smart inboxes are enabled for the
// installation/server.  This allows smart inboxes to be disabled per
// installation/database should performance become a problem
func SmartInboxesEnabled() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !viper.GetBool("smartinboxes.enabled") {
				return models.ErrDownForMaintenance
			}

			return next(c)
		}
	}
}
