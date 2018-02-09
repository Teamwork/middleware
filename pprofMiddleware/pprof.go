package pprofMiddleware

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"
)

// ProfileRequest creates a pprof CPU Profile for requests in dir.
func ProfileRequest(dir string) func(http.Handler) http.Handler {

	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
		}
		if err != nil {
			panic(err)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			fn := filepath.Join(dir, fmt.Sprintf(
				"%d-%s-%s", time.Now().Unix(), r.Method,
				strings.Replace(r.URL.Path, "/", ".", -1)))

			f, err := os.Create(fn)
			if err != nil {
				panic(err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			if err := pprof.StartCPUProfile(f); err != nil {
				panic(err)
			}
			defer pprof.StopCPUProfile()

			next.ServeHTTP(w, r)
		})
	}
}
