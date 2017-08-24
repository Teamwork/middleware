package middleware

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/teamwork/desk/helper"
	"github.com/teamwork/desk/models"
	"github.com/teamwork/desk/modules/httperr"
)

// CustomerPortalEnabled validates that the installation has access to the
// portal either through their billing plan or a trial use
func CustomerPortalEnabled() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := helper.GetSession(c)
			if err != nil {
				return httperr.WithCode(err, http.StatusNotFound)
			}
			enabled, err := models.IsEnabled("customerPortal", session)
			if err != nil {
				return errors.Wrap(err, "An error occurred looking up the customer portal.  Please try again")
			}
			if !enabled {
				return models.ErrBillingUpgradeRequired
			}
			return next(c)
		}
	}
}
