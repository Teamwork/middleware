package middleware

import (
	"github.com/labstack/echo"
	"github.com/spf13/viper"
)

func StaticRewriteIndex() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL().Path()
			if path == "" || path == "/index.html" {
				if viper.GetBool("static.useDist") {
					c.Request().URL().SetPath("/dist/index.html")
				}

				c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Response().Header().Set("Pragma", "no-cache")
				c.Response().Header().Set("Expires", "0")
			}

			return next(c)
		}
	}
}
