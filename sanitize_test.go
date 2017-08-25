package middleware // import "github.com/teamwork/middleware"

import (
	"net/url"
	"strings"
	"testing"
)

func Test_replaceIndexedFormFields(t *testing.T) {
	f := url.Values{}

	expected := map[string][]string{}
	expected["username"] = []string{"user"}
	expected["foods[]"] = []string{"cookies", "pizza"}

	f.Add("username", expected["username"][0])
	f.Add("foods[0]", expected["foods[]"][0])
	f.Add("foods[]", expected["foods[]"][1])

	f = replaceIndexedFormFields(f)
	if len(f) != len(expected) {
		t.Errorf("expected %d form keys, got %d", len(expected), len(f))
	}
	for key, values := range f {
		if strings.Contains(key, "[0]") {
			t.Error("expected indexed field to be removed")
		}

		if len(values) != len(expected[key]) {
			t.Errorf("expected key to contain %d values, got %d", len(expected[key]), len(values))
		}
	}
}
