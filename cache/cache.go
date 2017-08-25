package cache // import "github.com/teamwork/middleware/cache"

import "net/http"

// Disable sets the Cache-Control header to "no-cache" to prevent browsing
// caching.
func Disable(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		f(w, r)
	}
}
