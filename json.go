package middleware // import "github.com/teamwork/middleware"

import (
	"mime"
	"net/http"
)

// RequireJSON returns http.StatusUnsupportedMediaType if the request is not
// of type application/json.
func RequireJSON(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); ct != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte("Content-Type must be application/json"))
			return
		}

		f(w, r)
	}
}
