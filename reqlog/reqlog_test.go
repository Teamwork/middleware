package reqlog

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReqLog(t *testing.T) {
	buf := bytes.NewBufferString("")
	rr := HTTP(t, nil, Log(Options{File: buf})(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("handler"))
	}))

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: expected %v, got %v",
			http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if body != "handler" {
		t.Errorf("wrong body: expected %#v, got %#v",
			"body", body)
	}

	if !strings.Contains(buf.String(), "200") {
		t.Errorf("statuscode not in log line: %v", buf)
	}
}

// HTTP sets up a HTTP test. A GET request will be made for you if req is nil.
func HTTP(t *testing.T, req *http.Request, h http.HandlerFunc) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h)
	if req == nil {
		var err error
		req, err = http.NewRequest("GET", "", nil)
		if err != nil {
			t.Fatalf("cannot make request: %v", err)
		}
	}

	handler.ServeHTTP(rr, req)
	return rr
}
