// Package ratelimit implements rate limiting of HTTP requests.
package ratelimit // import "github.com/teamwork/middleware/ratelimit"

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/ripexz/rip"
)

var (
	// ErrInvalidRate is used when the rate is less than 1 request per second.
	ErrInvalidRate = errors.New("invalid rate; needs to be between 1 and 3600 seconds")

	// perPeriod is the number of API calls (to all endpoints) that can be made
	// by the client before receiving a 429 error
	perPeriod     = 20
	periodSeconds = 60

	// keyExpiration is how long the keys are kept in redis when no longer unused
	keyExpiration = 86400
)

// Helper function to make it easier to test.
var now = func() time.Time { return time.Now() }

type redisPool interface {
	Get() redis.Conn
}

type redisPoolCtx interface {
	GetWithContext(ctx context.Context) redis.Conn
}

// Config for RateLimit
type Config struct {
	Pool       redisPool
	GrantOnErr bool
	ErrorLog   func(error, string)

	// GetKey generates bucket keys.
	GetKey func(*http.Request) string

	// Ignore rate limit verification for this request if this returns true.
	Ignore func(*http.Request) bool

	// Rates returns the number of API calls (to all endpoints) that can be made
	// by the client considering the request. If null global values will be used.
	Rates func(*http.Request) (perPeriod, periodSeconds int)
}

// SetRate set the rate limit rate.
// (10, time.Second) is 10 requests per second
func SetRate(n int, d time.Duration) error {
	if d < time.Second || d > time.Hour {
		return ErrInvalidRate
	}

	perPeriod = n
	periodSeconds = int(d.Seconds())
	return nil
}

// IPBucket is a generator of rate limit buckets based on client's IP address.
// Optional filter function can be passed in, defaults to finding first public IP,
// see: https://godoc.org/github.com/ripexz/rip
func IPBucket(prefix string, filter func([]string) (string, bool)) func(*http.Request) string {
	return func(req *http.Request) string {
		return fmt.Sprintf("%s-%s", prefix, rip.FromRequest(req, filter))
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
			fmt.Fprintf(os.Stderr, "%v: %v", desc, err) // nolint: errcheck
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.Ignore != nil && opts.Ignore(r) {
				next.ServeHTTP(w, r)
				return
			}

			perPeriodLocal := perPeriod
			periodSecondsLocal := periodSeconds
			if opts.Rates != nil {
				perPeriodLocal, periodSecondsLocal = opts.Rates(r)
			}

			key := opts.GetKey(r)
			granted, remaining, err := grant(r.Context(), &opts, key, perPeriodLocal, periodSecondsLocal)
			if err != nil {
				opts.ErrorLog(err, "failed to check if access is granted")
				// returns an extra header when redis is down
				w.Header().Add("X-Rate-Limit-Err", "1")
				granted = opts.GrantOnErr
			}

			w.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(perPeriodLocal))
			w.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-Rate-Limit-Reset", strconv.Itoa(periodSecondsLocal))

			if !granted {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// grant checks if the access is granted for this bucket key.
var grant = func(
	ctx context.Context,
	opts *Config,
	key string,
	perPeriod, periodSeconds int,
) (granted bool, remaining int, err error) {

	accessTime := now().UnixNano()
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", periodSeconds))
	if err != nil {
		return false, 0, err
	}

	var conn redis.Conn
	if poolCtx, ok := opts.Pool.(redisPoolCtx); ok {
		// Optional approach to retrieve the pool injecting the request context.
		// Useful for encapsulating the pool layer for request specific behaviours
		// (e.g. DataDog tracer).
		conn = poolCtx.GetWithContext(ctx)
	} else {
		conn = opts.Pool.Get()
	}

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

	// Make sure to expire the key when it's no longer in use
	err = conn.Send("EXPIRE", key, keyExpiration)
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

	_, err = conn.Do("EXPIRE", key, 60*60*24*30)
	if err != nil {
		return false, 0, errors.Wrap(err, "failed to set TTL on zlist")
	}

	keys, err := redis.Strings(results[len(results)-1], err)
	if err != nil {
		return false, 0, errors.Wrap(err, "failed to parse results")
	}

	remaining = perPeriod - len(keys)

	// the reason why we consider remaining zero as valid is because we register
	// the current request before counting. For example, if we have 1 request
	// per minute allowed, even if I send a request every 60 seconds, the
	// remaining would be still zero.
	return remaining >= 0, remaining, nil
}
