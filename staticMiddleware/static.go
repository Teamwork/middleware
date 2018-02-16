// Package staticMiddleware contains some middlewares for static file routes.
package staticMiddleware // import "github.com/teamwork/middleware/staticMiddleware"

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// BlockTraversal prevents "../../../../../etc/passwd" type path traversal
// attacks for static routes: all paths must begin with root path.
func BlockTraversal(root string) func(http.Handler) http.Handler {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		panic(fmt.Errorf("cannot get absolute path for %v: %v", root, err))
	}
	root = rootAbs

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			abs, err := filepath.Abs(root + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "Not found: %v", r.URL.Path)
				return
			}

			if !strings.HasPrefix(abs, root) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BlockDotfiles prevents access to any files or directories that start with a
// dot (.).
//
// This includes the filename itself (e.g. ".gitignore") or any parent directory
// starting with a dit (e.g. ".git/foo/bar").
func BlockDotfiles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Run a simple check first; after that test it again on the absolute
		// path to be sure we're not blocking access to e.g. /../foo or /./foo.
		//
		// path.Abs is comparatively slow, and this makes this middleware about
		// twice as fast.
		if strings.Contains(r.URL.Path, "/.") {
			abs, _ := filepath.Abs(r.URL.Path)
			if strings.Contains(abs, "/.") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
