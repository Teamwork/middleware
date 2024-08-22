package rescue

import (
	"net/http"
	"testing"

	"github.com/teamwork/test"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))

}

func TestRescue(t *testing.T) {
	rr := test.HTTP(t, nil, Rescue(nil, false)(handle{}))

	if rr.Code != 200 {
		t.Errorf("want code %v, got %v", 200, rr.Code)
	}
	if b := rr.Body.String(); b != "handler" {
		t.Errorf("body wrong:\nwant: %#v\ngot:  %#v\n", "handler", b)
	}
}

type panicy struct{}

func (h panicy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("oh noes!")

}

func TestRescuePanic(t *testing.T) {
	rr := test.HTTP(t, nil, Rescue(nil, false)(panicy{}))

	if rr.Code != 500 {
		t.Errorf("want code %v, got %v", 200, rr.Code)
	}

	want := "Sorry, the server ran into a problem processing this request."
	if b := rr.Body.String(); b != want {
		t.Errorf("body wrong:\nwant: %#v\ngot:  %#v\n", want, b)
	}
}
