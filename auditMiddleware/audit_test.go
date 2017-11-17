package auditMiddleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/teamwork/test"
	"github.com/teamwork/test/diff"
)

type testAuditor struct {
	audit Audit
}

func (t *testAuditor) UserID(*http.Request) int64         { return 1 }
func (t *testAuditor) InstallationID(*http.Request) int64 { return 1 }
func (t *testAuditor) Log(*http.Request, error)           {}

func (t *testAuditor) AddAudit(r *http.Request, a *Audit) {
	t.audit = *a
}

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))

}

func TestAudit(t *testing.T) {
	cases := []struct {
		req  http.Request
		opts Options
		want Audit
	}{
		{
			req: http.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme:   "http",
					Host:     "example.com",
					Path:     "/foo",
					RawQuery: "x=xyz",
				},
				Form:       url.Values{"w00t": []string{"XXX"}},
				Header:     http.Header{"Some-Val": []string{"asd"}},
				ProtoMajor: 1,
				ProtoMinor: 1,
				Body:       ioutil.NopCloser(strings.NewReader("Hello")),
			},
			opts: Options{
				MaxSize: -1,
			},
			want: Audit{
				UserID:         1,
				InstallationID: 1,
				Path:           "/foo",
				Method:         "POST",
				RequestHeaders: Header{http.Header{
					"Some-Val": []string{"asd"},
				}},
				QueryParams: Values{url.Values{
					"x": []string{"xyz"},
				}},
				FormParams: Values{url.Values{
					"w00t": []string{"XXX"},
				}},
				RequestBody: []byte("Hello"),
			},
		},

		// MaxSize
		{
			req: http.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme:   "http",
					Host:     "example.com",
					Path:     "/foo",
					RawQuery: "x=xyz",
				},
				Form:       url.Values{"w00t": []string{"XXX"}, "asd": []string{"zxcv", "qweqwe"}},
				Header:     http.Header{"Some-Val": []string{"asd"}},
				ProtoMajor: 1,
				ProtoMinor: 1,
				Body:       ioutil.NopCloser(strings.NewReader("Hello")),
			},
			opts: Options{
				MaxSize: 1,
			},
			want: Audit{
				UserID:         1,
				InstallationID: 1,
				Path:           "/foo",
				Method:         "POST",
				RequestHeaders: Header{http.Header{
					"Some-Val": []string{"asd"},
				}},
				QueryParams: Values{url.Values{
					"x": []string{"xyz"},
				}},
				FormParams: Values{url.Values{
					"w00t": []string{"X"},
					"asd":  []string{"z", "q"},
				}},
				RequestBody: []byte("H"),
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			a := testAuditor{}
			tc.opts.Auditor = &a

			rr := test.HTTP(t, &tc.req, Do(tc.opts)(handle{}).ServeHTTP)

			if rr.Code != 200 {
				t.Fatalf("code != 200: %v", rr.Code)
			}
			body := rr.Body.String()
			if body != "handler" {
				t.Fatalf("body != handler: %v", body)
			}

			a.audit.CreatedAt = Time{}
			if d := diff.Diff(tc.want, a.audit); d != "" {
				t.Errorf(d)
			}
		})
	}
}
