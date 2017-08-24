package middleware

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"github.com/teamwork/log"
)

// Path prevents "../../../../../etc/passwd" type path traversal attacks for
// static routes: all paths must begin with DESKPATH.
func Path() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		root := viper.GetString("deskpath")
		return func(c echo.Context) error {
			abs, err := filepath.Abs(root + "/" + c.Request().URI())
			if err != nil {
				log.Errorf(err, "error getting absolute path")
				return c.NoContent(http.StatusInternalServerError)
			}
			if !strings.HasPrefix(abs, root) {
				return c.NoContent(http.StatusForbidden)
			}
			return next(c)
		}
	}
}
