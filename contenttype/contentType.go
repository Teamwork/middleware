// Package contenttype validates the Content-Type header.
package contenttype // import "github.com/teamwork/middleware/contenttype"

import (
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/teamwork/utils/sliceutil"
)

// Options for Validate().
type Options struct {
	// Methods to validate; if the request method isn't in the list we won't do
	// anything.
	Methods []string

	// List of valid content-types.
	ValidContentTypes []string

	// AllowEmpty indicates if not sending a Content-Type header is allowed.
	AllowEmpty bool
}

// Some commonly used Content-Type values.
const (
	ContentTypeJSON        = "application/json"
	ContentTypeFormEncoded = "application/x-www-form-urlencoded"
	ContentTypeFormData    = "multipart/form-data"
)

// DefaultOptions for the Valid() middleware if options if nil.
var DefaultOptions = &Options{
	Methods:           []string{http.MethodPost, http.MethodPut, http.MethodDelete},
	ValidContentTypes: []string{ContentTypeJSON, ContentTypeFormEncoded, ContentTypeFormData},
	AllowEmpty:        false,
}

// JSON are options to allow only JSON for all requests verbs (GET included).
var JSON = &Options{
	Methods:           []string{},
	ValidContentTypes: []string{ContentTypeJSON},
	AllowEmpty:        false,
}

// Validate ensures that the content type header is set and is one of the
// allowed content types. This ONLY applies to POST, PUT, and DELETE requests.
func Validate(opts *Options) func(http.Handler) http.Handler {
	if opts == nil {
		opts = DefaultOptions
	}
	if len(opts.ValidContentTypes) == 0 {
		panic("no valid Content-Type given")
	}
	for i := range opts.Methods {
		opts.Methods[i] = strings.ToUpper(opts.Methods[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(opts.Methods) > 0 && !sliceutil.InStringSlice(opts.Methods, r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			if opts.AllowEmpty && r.Header.Get("Content-Type") == "" {
				next.ServeHTTP(w, r)
				return
			}

			ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if ct == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(fmt.Sprintf("invalid content type: %v",
					r.Header.Get("Content-Type"))))
				return
			}

			for _, valid := range opts.ValidContentTypes {
				if ct == valid {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("unknown content type: %v; must be one of %v",
				ct, strings.Join(opts.ValidContentTypes, ", "))))
		})
	}
}
