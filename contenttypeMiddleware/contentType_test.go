package contenttypeMiddleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/teamwork/test"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))

}

func TestValidate(t *testing.T) {
	cases := []struct {
		in       *http.Request
		wantBody string
		wantCode int
	}{
		{
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"woot woot"},
				},
			},
			"handler",
			http.StatusOK,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"woot woot"},
				},
			},
			"invalid content type: woot woot",
			http.StatusBadRequest,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			"handler",
			http.StatusOK,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/JSON; encoding=sd"},
				},
			},
			"handler",
			http.StatusOK,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{""},
				},
			},
			"invalid content type: ",
			http.StatusBadRequest,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"So this isn't really a valid value"},
				},
			},
			"invalid content type: So this isn't really a valid value",
			http.StatusBadRequest,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/woot"},
				},
			},
			"unknown content type: application/woot; " +
				"must be one of application/json, application/x-www-form-urlencoded, multipart/form-data",
			http.StatusBadRequest,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/jsonEXTRA"},
				},
			},
			"unknown content type: application/jsonextra; " +
				"must be one of application/json, application/x-www-form-urlencoded, multipart/form-data",
			http.StatusBadRequest,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			rr := test.HTTP(t, tc.in, Validate(nil)(handle{}).ServeHTTP)

			if rr.Code != tc.wantCode {
				t.Errorf("want code %v, got %v", tc.wantCode, rr.Code)
			}
			if b := rr.Body.String(); b != tc.wantBody {
				t.Errorf("body wrong:\nwant: %#v\ngot:  %#v\n", tc.wantBody, b)
			}
		})
	}
}

func TestValidateOptions(t *testing.T) {
	cases := []struct {
		in       *http.Request
		opts     *Options
		wantCode int
	}{
		{
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"woot/woot"},
				},
			},
			&Options{
				Methods:           []string{"GET"},
				ValidContentTypes: []string{"woot/woot"},
			},
			http.StatusOK,
		},
		{
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"woot"},
				},
			},
			&Options{
				Methods:           []string{"GET"},
				ValidContentTypes: []string{"woot/woot"},
			},
			http.StatusBadRequest,
		},
		{
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"woot"},
				},
			},
			JSON,
			http.StatusBadRequest,
		},
		{
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			JSON,
			http.StatusOK,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{},
			},
			&Options{
				Methods:           []string{http.MethodPost, http.MethodPut, http.MethodDelete},
				ValidContentTypes: []string{ContentTypeJSON, ContentTypeFormEncoded, ContentTypeFormData},
				AllowEmpty:        true,
			},
			http.StatusOK,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			rr := test.HTTP(t, tc.in, Validate(tc.opts)(handle{}).ServeHTTP)

			if rr.Code != tc.wantCode {
				t.Errorf("want code %v, got %v", tc.wantCode, rr.Code)
			}
		})
	}
}
