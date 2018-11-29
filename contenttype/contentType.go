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
var DefaultOptions = map[string]Options{
	"": Options{
		Methods:           []string{http.MethodPost, http.MethodPut, http.MethodDelete},
		ValidContentTypes: []string{ContentTypeJSON, ContentTypeFormEncoded, ContentTypeFormData},
		AllowEmpty:        false,
	},
}

// JSON are options to allow only JSON for all requests verbs (GET included).
var JSON = map[string]Options{
	"": Options{
		Methods:           []string{},
		ValidContentTypes: []string{ContentTypeJSON},
		AllowEmpty:        false,
	},
}

// Validate ensures that the content type header is set and is one of the
// allowed content types. This ONLY applies to POST, PUT, and DELETE requests.
//
// The option key is the pathname to check; use an empty string for the default
// options (this is mandatory).
func Validate(opts map[string]Options) func(http.Handler) http.Handler {
	if opts == nil {
		opts = DefaultOptions
	}
	for k := range opts {
		if len(opts[k].ValidContentTypes) == 0 {
			panic(fmt.Sprintf("no valid Content-Type given for %q", k))
		}
		for i := range opts[k].Methods {
			opts[k].Methods[i] = strings.ToUpper(opts[k].Methods[i])
		}
	}
	if _, ok := opts[""]; !ok {
		panic("no default options given")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			opt, ok := opts[r.URL.Path]
			if !ok {
				opt = opts[""]
			}

			if len(opt.Methods) > 0 && !sliceutil.InStringSlice(opt.Methods, r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			if opt.AllowEmpty && r.Header.Get("Content-Type") == "" {
				next.ServeHTTP(w, r)
				return
			}

			ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if ct == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(fmt.Sprintf("invalid content type: %q",
					r.Header.Get("Content-Type"))))
				return
			}

			for _, valid := range opt.ValidContentTypes {
				if ct == valid {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf(
				"unknown content type: %q for %q; must be one of %v",
				ct, r.URL.Path, strings.Join(opt.ValidContentTypes, ", "))))
		})
	}
}
