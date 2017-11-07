package ratelimitMiddleware // import "github.com/teamwork/middleware/ratelimitMiddleware"

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/teamwork/log"
)

const (
	// ErrRateLimit is used when the rate limit is reached and requests are
	// being throttled.
	ErrRateLimit = "Rate limit exceeded"

	// perPeriod is the number of API calls (to all endpoints) that can be made
	// by the client before receiving a 429 error
	perPeriod     int64 = 20
	periodSeconds int64 = 60
)

type rateLimiter struct {
	Remaining int64
	Limit     int64
	Reset     int64

	key     string
	tracker *redis.Pool
}

// Limit limits API requests based on the IP address
func Limit(p *redis.Pool) func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rateLimit := newRateLimiter(getKey(r), p)
			w.Header().Add("X-Rate-Limit-Limit", strconv.FormatInt(rateLimit.Limit, 10))
			w.Header().Add("X-Rate-Limit-Remaining", strconv.FormatInt(rateLimit.Remaining, 10))
			w.Header().Add("X-Rate-Limit-Reset", strconv.FormatInt(rateLimit.Reset, 10))

			if rateLimit.limitIsReached() {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			f(w, r)
		}
	}
}

func getKey(r *http.Request) string {
	h := fnv.New32a()
	u, err := url.Parse(r.RequestURI)
	if err != nil {
		h.Write([]byte(r.RequestURI)) // nolint: errcheck
	} else {
		h.Write([]byte(u.Host)) // nolint: errcheck
	}

	return fmt.Sprintf("%v-%v", "", h.Sum32())
}

func newRateLimiter(key string, pool *redis.Pool) rateLimiter {
	limiter := rateLimiter{
		key:     fmt.Sprintf("launchpad-rate-limiter-%s", key),
		tracker: pool,
		Limit:   perPeriod,
		Reset:   periodSeconds,
	}

	err := limiter.appendRequest()
	if err != nil {
		log.Error(err, "failed to append request")
	}

	return limiter
}

func (rl *rateLimiter) limitIsReached() bool {
	return rl.Remaining < 1
}

func (rl *rateLimiter) appendRequest() error {
	accessTime := time.Now().UnixNano()
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", periodSeconds))
	if err != nil {
		return err
	}

	conn := rl.tracker.Get()
	defer conn.Close() // nolint: errcheck

	err = conn.Send("MULTI")
	if err != nil {
		return err
	}

	// Add the new request to the bucket
	err = conn.Send("ZADD", rl.key, accessTime, accessTime)
	if err != nil {
		return err
	}

	// Remove any keys that are outside of the interval
	err = conn.Send("ZREMRANGEBYSCORE", rl.key, 0, accessTime-duration.Nanoseconds())
	if err != nil {
		return err
	}

	// Check how many keys we have in the set
	err = conn.Send("ZRANGE", rl.key, 0, -1)
	if err != nil {
		return err
	}

	results, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	keys, err := redis.Strings(results[len(results)-1], err)
	if err != nil {
		return errors.Wrap(err, "failed to parse results")
	}

	rl.Remaining = rl.Limit - int64(len(keys))

	return nil
}
