// Package ratelimitMiddleware implements rate limiting of HTTP requests.
// The default rate is 20 requests per minute. This rate can be changed using SetRate.
package ratelimitMiddleware // import "github.com/teamwork/middleware/ratelimitMiddleware"

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/tomasen/realip"
)

const (
	// ErrInvalidRate is used when the rate is less than 1 request per second.
	ErrInvalidRate = "Invalid rate"
)

var (
	// perPeriod is the number of API calls (to all endpoints) that can be made
	// by the client before receiving a 429 error
	perPeriod     = 20
	periodSeconds = 60
)

// Config for RateLimit
type Config struct {
	Pool       *redis.Pool
	GrantOnErr bool
	ErrorLog   func(error, string)

	// GetKey generates bucket keys.
	GetKey func(req *http.Request) string

	// Ignore rate limit verification if this returns true.
	Ignore func(req *http.Request) bool
}

// SetRate set the rate limit rate.
// (10, time.Second) is 10 requests per second
func SetRate(n int, d time.Duration) error {
	if d < time.Second || d > time.Hour {
		return errors.New(ErrInvalidRate)
	}

	perPeriod = n
	periodSeconds = int(d.Seconds())
	return nil
}

// IPBucket is a generator of rate limit buckets based on client's IP address
func IPBucket(prefix string, req *http.Request) func(req *http.Request) string {
	return func(req *http.Request) string {
		return fmt.Sprintf("%s-%s", prefix, realip.RealIP(req))
	}
}

// RateLimit limits requests for a key provided by getKey function.
// If ignore function returns true, rate limit is bypassed.
// grantOnErr argument defines if it should grant access when Redis is down.
func RateLimit(opts Config) func(http.Handler) http.Handler {

	if opts.GetKey == nil {
		panic("opts.GetKey is nil")
	}

	if opts.ErrorLog == nil {
		opts.ErrorLog = func(err error, desc string) {
			fmt.Fprintf(os.Stderr, "%v: %v", desc, err)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.Ignore != nil && opts.Ignore(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := opts.GetKey(r)
			granted, remaining, err := grant(&opts, key)
			if err != nil {
				opts.ErrorLog(err, "failed to check if access is granted")
				// returns an extra header when redis is down
				w.Header().Add("X-Rate-Limit-Err", "1")
				granted = opts.GrantOnErr
			}

			w.Header().Add("X-Rate-Limit-Limit", strconv.Itoa(perPeriod))
			w.Header().Add("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
			w.Header().Add("X-Rate-Limit-Reset", strconv.Itoa(periodSeconds))

			if !granted {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// grant checks if the access is granted for this bucket key
var grant = func(opts *Config, key string) (granted bool, remaining int, err error) {
	accessTime := now().UnixNano()
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", periodSeconds))
	if err != nil {
		return false, 0, err
	}

	conn := opts.Pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			opts.ErrorLog(err, "error when closing Redis connection")
		}
	}()

	err = conn.Send("MULTI")
	if err != nil {
		return false, 0, err
	}

	// Add the new request to the bucket
	err = conn.Send("ZADD", key, accessTime, accessTime)
	if err != nil {
		return false, 0, err
	}

	// Remove any keys that are outside of the interval
	err = conn.Send("ZREMRANGEBYSCORE", key, 0, accessTime-duration.Nanoseconds())
	if err != nil {
		return false, 0, err
	}

	// Check how many keys we have in the set
	err = conn.Send("ZRANGE", key, 0, -1)
	if err != nil {
		return false, 0, err
	}

	results, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return false, 0, errors.Wrap(err, "transaction failed")
	}

	keys, err := redis.Strings(results[len(results)-1], err)
	if err != nil {
		return false, 0, errors.Wrap(err, "failed to parse results")
	}

	remaining = perPeriod - len(keys)
	return remaining >= 1, remaining, nil
}

// a helper function to make it easier to test
var now = func() time.Time {
	return time.Now()
}
