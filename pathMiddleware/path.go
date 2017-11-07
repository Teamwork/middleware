// Package pathMiddleware prevents path traversal attacks.
package pathMiddleware // import "github.com/teamwork/middleware/pathMiddleware"

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/teamwork/log"
)

// BlockTraversal prevents "../../../../../etc/passwd" type path traversal
// attacks for static routes: all paths must begin with root path.
func BlockTraversal(root string) func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			rootAbs, err := filepath.Abs(root)
			if err != nil {
				log.Errorf(err, "error getting absolute path")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			abs, err := filepath.Abs(root + r.RequestURI)
			if err != nil {
				log.Errorf(err, "error getting absolute path")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if !strings.HasPrefix(abs, rootAbs) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			f(w, r)
		}
	}
}
