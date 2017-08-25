package reqlog // import "github.com/teamwork/middleware/reqlog"

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/teamwork/log"
)

// Log requests to stdout. This is mainly intended for dev-env.
func Log(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)

		// Don't log health requests since there are so many!
		if r.RequestURI == "/health.json" {
			return
		}

		statusCode, err := strconv.ParseInt(w.Header().Get("status"), 10, 64)
		if err != nil {
			return
		}

		status := " "
		switch {
		case statusCode >= 200 && statusCode < 400:
			status = "\x1b[48;5;154m\x1b[38;5;0m%v\x1b[0m"
		case statusCode >= 400 && statusCode <= 499:
			status = "\x1b[1m\x1b[48;5;221m\x1b[38;5;0m%v\x1b[0m"
		case statusCode >= 500 && statusCode <= 599:
			status = "\x1b[1m\x1b[48;5;9m\x1b[38;5;15m%v\x1b[0m"
		}

		fmt.Printf("%v %v %v   %v%v\n",
			time.Now().Format(log.TimeFormat),
			fmt.Sprintf(status, statusCode), r.Method, r.Host,
			r.RequestURI)

		return
	}
}
