// Package cache disables browser caching.
//
// See: https://tools.ietf.org/html/rfc7234#section-5.2.1.4
package cache // import "github.com/teamwork/middleware/cache"

import "net/http"

// NoCache sets the Cache-Control header to "no-cache". This tells browsers to
// always validate a cache (with e.g. If-Match or If-None-Match). It does NOT
// tell browsers to never store a cache (use NoStore for that).
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

// NoStore sets the Cache-Control header to "no-store, no-cache" which tells
// browsers to never store a local copy (the no-cache is there to be sure
// previously stored copies from before this header are revalidated).
func NoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store,no-cache")
		next.ServeHTTP(w, r)
	})
}
