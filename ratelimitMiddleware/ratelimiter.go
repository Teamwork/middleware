// Package ratelimitMiddleware implements rate limiting of HTTP requests.
// The default rate is 20 requests per minute. This rate can be changed using SetRate.
package ratelimitMiddleware // import "github.com/teamwork/middleware/ratelimitMiddleware"

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/teamwork/log"
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

// SetRate set the rate limit rate.
// (10, time.Second) is 10 requests per second
func SetRate(n int, d time.Duration) error {
	if d < time.Second && d > time.Hour {
		return errors.New(ErrInvalidRate)
	}

	perPeriod = n
	periodSeconds = int(d.Seconds())
	return nil
}

// GetKeyFunc is a function that generates bucket keys.
type GetKeyFunc func(req *http.Request) string

// IgnoreFunc determines when rate limit verification should be ignored.
type IgnoreFunc func(req *http.Request) bool

// IPBucket is a generator of rate limit buckets based on
// client's IP address
func IPBucket(prefix string, req *http.Request) GetKeyFunc {
	return func(req *http.Request) string {
		return fmt.Sprintf("%s-%s", prefix, realip.RealIP(req))
	}
}

// RateLimit limits requests for a key provided by getKey function.
// If ignore function returns true, rate limit is bypassed.
func RateLimit(p *redis.Pool, getKey GetKeyFunc, ignore IgnoreFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ignore != nil && ignore(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := getKey(r)
			granted, remaining, err := grant(p, key)
			if err != nil {
				log.Error(err, "failed to check if access is granted")
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
var grant = func(pool *redis.Pool, key string) (granted bool, remaining int, err error) {
	accessTime := now().UnixNano()
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", periodSeconds))
	if err != nil {
		return false, 0, err
	}

	conn := pool.Get()
	defer conn.Close() // nolint: errcheck

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
