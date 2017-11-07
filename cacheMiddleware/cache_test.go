package cacheMiddleware

import (
	"net/http"
	"testing"

	"github.com/teamwork/test"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))

}

func TestNoCache(t *testing.T) {
	rr := test.HTTP(t, nil, NoCache(handle{}).ServeHTTP)
	if rr.Code != http.StatusOK {
		t.Errorf("want code 200, got %v", rr.Code)
	}
	if b := rr.Body.String(); b != "handler" {
		t.Errorf("body wrong: %#v", b)
	}
	if h := rr.Header().Get("Cache-Control"); h != "no-cache" {
		t.Errorf("header wrong: %#v", h)
	}
}

func TestNoStore(t *testing.T) {
	rr := test.HTTP(t, nil, NoStore(handle{}).ServeHTTP)

	if rr.Code != http.StatusOK {
		t.Errorf("want code 200, got %v", rr.Code)
	}
	if b := rr.Body.String(); b != "handler" {
		t.Errorf("body wrong: %#v", b)
	}
	if h := rr.Header().Get("Cache-Control"); h != "no-store,no-cache" {
		t.Errorf("header wrong: %#v", h)
	}
}
