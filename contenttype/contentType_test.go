package contenttype

import (
	"fmt"
	"net/http"
	"net/url"
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
			`invalid content type: "woot woot"`,
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
			`invalid content type: ""`,
			http.StatusBadRequest,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"So this isn't really a valid value"},
				},
			},
			`invalid content type: "So this isn't really a valid value"`,
			http.StatusBadRequest,
		},
		{
			&http.Request{
				URL:    &url.URL{Path: "/p"},
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/woot"},
				},
			},
			`unknown content type: "application/woot" for "/p"; ` +
				"must be one of application/json, application/x-www-form-urlencoded, multipart/form-data",
			http.StatusBadRequest,
		},
		{
			&http.Request{
				URL:    &url.URL{Path: "/p"},
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/jsonEXTRA"},
				},
			},
			`unknown content type: "application/jsonextra" for "/p"; ` +
				"must be one of application/json, application/x-www-form-urlencoded, multipart/form-data",
			http.StatusBadRequest,
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			if tt.in.URL == nil {
				tt.in.URL = &url.URL{Path: "/"}
			}
			rr := test.HTTP(t, tt.in, Validate(nil)(handle{}))

			if rr.Code != tt.wantCode {
				t.Errorf("want code %v, got %v", tt.wantCode, rr.Code)
			}
			if b := rr.Body.String(); b != tt.wantBody {
				t.Errorf("body wrong:\nwant: %#v\ngot:  %#v\n", tt.wantBody, b)
			}
		})
	}
}

func TestValidateOptions(t *testing.T) {
	cases := []struct {
		in       *http.Request
		opts     map[string]Options
		wantCode int
	}{
		{
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"woot/woot"},
				},
			},
			map[string]Options{"": {
				Methods:           []string{"GET"},
				ValidContentTypes: []string{"woot/woot"},
			}},
			http.StatusOK,
		},
		{
			&http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"woot"},
				},
			},
			map[string]Options{"": {
				Methods:           []string{"GET"},
				ValidContentTypes: []string{"woot/woot"},
			}},
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
			map[string]Options{"": {
				Methods:           []string{http.MethodPost, http.MethodPut, http.MethodDelete},
				ValidContentTypes: []string{ContentTypeJSON, ContentTypeFormEncoded, ContentTypeFormData},
				AllowEmpty:        true,
			}},
			http.StatusOK,
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			tt.in.URL = &url.URL{Path: "/"}
			rr := test.HTTP(t, tt.in, Validate(tt.opts)(handle{}))

			if rr.Code != tt.wantCode {
				t.Errorf("want code %v, got %v", tt.wantCode, rr.Code)
			}
		})
	}
}

func TestPathMatch(t *testing.T) {
	cases := []struct {
		in       *http.Request
		opts     map[string]Options
		wantCode int
	}{
		{
			&http.Request{
				URL:    &url.URL{Path: "/foo"},
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"woot/woot"},
				},
			},
			map[string]Options{
				"": {
					Methods:           []string{"GET"},
					ValidContentTypes: []string{"asd/def"},
				},
				"/foo": {
					Methods:           []string{"GET"},
					ValidContentTypes: []string{"woot/woot"},
				},
			},
			http.StatusOK,
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			rr := test.HTTP(t, tt.in, Validate(tt.opts)(handle{}))
			if rr.Code != tt.wantCode {
				t.Errorf("want code %v, got %v", tt.wantCode, rr.Code)
			}
		})
	}
}
