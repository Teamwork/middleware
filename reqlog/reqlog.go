package reqlog // import "github.com/teamwork/middleware/reqlog"

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/teamwork/log"
)

// Options for the Log middleware.
type Options struct {
	// LogStart indicates that a log entry should be added both before and after
	// running the request. The default is to add a log entry only at the end.
	LogStart bool

	// LogID indicates that a random ID should be added to logs. This is
	// especially useful in combination with LogStart.
	LogID bool

	// SkipURL skips logging for this URL. Useful for health checks.
	SkipURL string
}

// Log requests to stdout.
func Log(opt Options) {
	var end string
	if opt.LogStart {
		end = "end   "
	}

	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Don't log health requests since there are so many!
			if opt.SkipURL != "" && r.URL.Path == opt.SkipURL {
				f(w, r)
				return
			}

			var rand string
			if opt.LogID {
				rand = makeRandom()
			}

			if opt.LogStart {
				fmt.Printf("start %v %v%v %v   %v%v\n",
					end, time.Now().Format(log.TimeFormat), rand, r.Method,
					r.Host, r.RequestURI)
			}

			f(w, r)

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

			fmt.Printf("%v%v%v %v %v   %v%v\n",
				end,
				time.Now().Format(log.TimeFormat), rand,
				fmt.Sprintf(status, statusCode), r.Method, r.Host,
				r.RequestURI)

			return
		}
	}
}

// makeRandom returns a 32bit random number as a hex string.
func makeRandom() string {
	uuid := make([]byte, 4)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}

	return fmt.Sprintf(" %x", uuid[0:4])
}
