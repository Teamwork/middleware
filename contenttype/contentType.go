package contenttype // import "github.com/teamwork/middleware/contenttype"

import (
	"mime"
	"net/http"
	"strings"
)

var validContentTypes = []string{
	"application/json",
	"application/x-www-form-urlencoded",
	"multipart/form-data",
}

// Validate ensures that the content type header is set and is one of
// the allowed content types.  This ONLY applies to POST and PUT.
func Validate(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToUpper(r.Method)
		if method != "POST" && method != "PUT" {
			f(w, r)
			return
		}

		ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if ct == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid content type"))
			return
		}

		for _, valid := range validContentTypes {
			if strings.Contains(ct, valid) {
				f(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid content type"))
	}
}
