package httphandlers

import (
	"mime"
	"net/http"
	"sort"

	"github.com/golang/gddo/httputil/header"
	"github.com/labstack/echo"

	"github.com/teamwork/httperr"
	"github.com/teamwork/log"
)

const (
	typeJSON = "application/json"
	typeText = "text/plain"
	typeHTML = "text/html"
)

// errType returns the content type that should be used for the error
func errType(req *http.Request, res http.ResponseWriter) string {
	// If we already committed a CT to return, use that
	if ct, _, _ := mime.ParseMediaType(res.Header().Get("Content-Type")); ct != "" {
		return ct
	}
	specs := header.ParseAccept(req.Header, "Accept")
	sort.Slice(specs, func(i, j int) bool { return specs[i].Q > specs[j].Q })
	for _, spec := range specs {
		switch ct, _, _ := mime.ParseMediaType(spec.Value); ct {
		case typeJSON, typeText, typeHTML:
			return ct
		case "text/*":
			return typeText
		}
	}
	return typeJSON
}

// ErrorHandler handles any errors that propogate back to the top level. It logs
// them, and presents an error page to the visitor.
func ErrorHandler(err error, c echo.Context) {
	code, msg := codeAndMessage(err)

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD { // Issue #608
			_ = c.NoContent(code)
		} else {
			switch ct := errType(c.Request(), c.Response().Writer); ct {
			case typeJSON:
				_ = c.JSON(code, map[string]interface{}{
					"status": "error",
					"errors": []string{msg},
				})
			default:
				_ = c.String(code, msg)
			}
		}
	}

	switch code {
	case http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusInsufficientStorage:
		log.WithHTTPRequest(c.Request()).Err(err)
	default:
		log.Printf("Path '%s': %s\n", c.Request().RequestURI, err.Error())
	}
}

// codeAndMessage extracts the HTTP status code and message from an error.
func codeAndMessage(err error) (statusCode int, msg string) {
	var message string
	if he, ok := err.(*echo.HTTPError); ok {
		statusCode = he.Code
		message = he.Message.(string)
	} else {
		statusCode = httperr.StatusCode(err)
		message = err.Error()
	}
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	return statusCode, message
}
