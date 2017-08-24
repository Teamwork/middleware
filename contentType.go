package middleware

import (
	"mime"
	"strings"

	"github.com/labstack/echo"
	"github.com/teamwork/log"
)

var validContentTypes = []string{
	"application/json",
	"application/x-www-form-urlencoded",
	"multipart/form-data",
}

// ValidateContentType ensures that the content type header is set and is one of
// the allowed content types.  This ONLY applies to POST and PUT.
func ValidateContentType() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			method := strings.ToLower(c.Request().Method())
			if method != "post" && method != "put" {
				return next(c)
			}

			ct, _, _ := mime.ParseMediaType(c.Request().Header().Get("Content-Type"))
			for _, valid := range validContentTypes {
				// We use valid as content type headers can additionally
				// specify the encoding
				if strings.Contains(ct, valid) {
					return next(c)
				}
			}

			log.WithFields(log.Fields{
				"URI":    c.Request().URI(),
				"Method": method,
			}).Println("content-type header not set")
			return next(c)
		}
	}
}
