package middleware // import "github.com/teamwork/middleware"

import "net/http"

// NoCache sets the Cache-Control header to "no-cache" to prevent browsing
// caching.
func NoCache(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		f(w, r)
	}
}
