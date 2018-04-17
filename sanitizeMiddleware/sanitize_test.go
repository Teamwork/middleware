package sanitize

import (
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
)

func Test_StripFileExtensionWithJson(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/teamwork.json", nil)
	res := test.NewResponseRecorder()
	c := e.NewContext(req, res)
	c.SetPath("/:id.json")
	c.SetParamNames("id")
	c.SetParamValues("teamwork")

	_ = StripFileExtensions()(func(echo.Context) error {
		val := c.ParamValues()[0]

		if val != "teamwork" {
			t.Error("")
		}

		return nil
	})(c)
}

func Test_StripFileExtensionWithoutJson(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/teamwork_no_json", nil)
	res := test.NewResponseRecorder()
	c := e.NewContext(req, res)
	c.SetPath("/:id.json")
	c.SetParamNames("id")
	c.SetParamValues("teamwork")

	_ = StripFileExtensions()(func(echo.Context) error {
		val := c.ParamValues()[0]

		if val != "teamwork_no_json" {
			t.Error("")
		}

		return nil
	})(c)
}
