package httperr

import (
	"net/http"
	"testing"
)

func TestStdErrError(t *testing.T) {
	t.Run("ValidStatus", func(t *testing.T) {
		var err error = ErrGatewayTimeout
		msg := err.Error()
		expected := http.StatusText(http.StatusGatewayTimeout)
		if msg != expected {
			t.Errorf("Expected '%s', got '%s'", expected, msg)
		}
	})

	t.Run("UnknownStatus", func(t *testing.T) {
		var err error = stdErr(9000)
		msg := err.Error()
		expected := "Unknown Status"
		if msg != expected {
			t.Errorf("Expected '%s', got '%s'", expected, msg)
		}
	})
}

func TestStdErrStatusCode(t *testing.T) {
	var err error = ErrGatewayTimeout
	if code := StatusCode(err); code != http.StatusGatewayTimeout {
		t.Errorf("Expected %d, got %d", http.StatusGatewayTimeout, code)
	}
}
