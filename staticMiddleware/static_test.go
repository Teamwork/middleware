package staticMiddleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teamwork/test"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))
}

func TestBlockTraversal(t *testing.T) {
	cases := []struct {
		root, request string
		wantCode      int
	}{
		{"/root", "/some/path", 200},
		{"/root", "/some/path.ext", 200},
		{"/root", "/some/../path..ext", 200},
		{"/root", "/some/../path", 200},
		{"/root", "/some/../../path", 403},
		{"/root", "/some/./././path", 200},
		{"/root", "/some/\\/././path", 200},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.request, nil)
			if err != nil {
				t.Fatalf("cannot make request: %v", err)
			}

			rr := test.HTTP(t, req, BlockTraversal(tc.root)(handle{}).ServeHTTP)
			if rr.Code != tc.wantCode {
				t.Errorf("\nout:  %#v\nwant: %#v\n", rr.Code, tc.wantCode)
			}
			if rr.Code != 200 && rr.Body.String() != "" {
				t.Errorf("expected body to be empty: %#v", rr.Body.String())
			}
		})
	}
}

func TestBlockDotfiles(t *testing.T) {
	cases := []struct {
		request  string
		wantCode int
	}{
		{"/some/path", 200},
		{"/some/path.", 200},
		{"/some/path.ext", 200},
		{"/some/../../path", 200},
		{"/some/./path", 200},
		{"/some/path/.hidden", 404},
		{"/some/path/.hidden/file", 404},
		{"/.some/path/hidden/file", 404},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.request, nil)
			if err != nil {
				t.Fatalf("cannot make request: %v", err)
			}

			rr := test.HTTP(t, req, BlockDotfiles(handle{}).ServeHTTP)
			if rr.Code != tc.wantCode {
				t.Errorf("\nout:  %#v\nwant: %#v\n", rr.Code, tc.wantCode)
			}
			if rr.Code != 200 && rr.Body.String() != "" {
				t.Errorf("expected body to be empty: %#v", rr.Body.String())
			}
		})
	}
}

func BenchmarkBlockDotfiles(b *testing.B) {
	f := BlockDotfiles(handle{}).ServeHTTP
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/some/path/to/a/file", nil)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		f(w, r)
	}
}
