package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/teamwork/desk/helper"
	"github.com/teamwork/desk/modules/cache"
	"github.com/teamwork/log"
)

var (
	perPeriod     int64 = 120
	periodSeconds int64 = 60
)

const (
	// ErrRateLimit is used when the rate limit is reached and requests are
	// being throttled.
	ErrRateLimit = "Rate limit exceeded"
)

type rateLimiter struct {
	Remaining int64
	Limit     int64
	Reset     int64

	key     string
	tracker redis.Conn
}

// RateLimit limits API requests based on the user ID.
func RateLimit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header().Get("DeskWeb") != "" {
				return next(c)
			}

			session, err := helper.GetSession(c)
			if err != nil {
				log.Error(err, "failed to get session")
				return next(c)
			}

			rateLimit := newRateLimiter(session.ID)
			c.Response().Header().Add("X-Rate-Limit-Limit", strconv.FormatInt(rateLimit.Limit, 10))
			c.Response().Header().Add("X-Rate-Limit-Remaining", strconv.FormatInt(rateLimit.Remaining, 10))
			c.Response().Header().Add("X-Rate-Limit-Reset", strconv.FormatInt(rateLimit.Reset, 10))

			if rateLimit.limitIsReached() {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"errors": []map[string]interface{}{
						{
							"code":    http.StatusTooManyRequests,
							"message": ErrRateLimit,
						},
					},
				})
			}

			return next(c)
		}
	}
}

func newRateLimiter(key int64) rateLimiter {
	limiter := rateLimiter{
		key:     fmt.Sprintf("rate-limiter-%d", key),
		tracker: cache.GetDeskDataConnection(),
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
	if !viper.GetBool("redis.enabled") {
		// Mocking this is a pain; for now just disable rate limiting if
		// redis is off.
		rl.Remaining = 1
		return nil
	}
	accessTime := time.Now().UnixNano()
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", periodSeconds))
	if err != nil {
		return err
	}

	rl.tracker.Send("MULTI")

	// Add the new request to the bucket
	rl.tracker.Send("ZADD", rl.key, accessTime, accessTime)

	// Remove any keys that are outside of the interval
	rl.tracker.Send("ZREMRANGEBYSCORE", rl.key, 0, accessTime-duration.Nanoseconds())

	// Check how many keys we have in the set
	rl.tracker.Send("ZRANGE", rl.key, 0, -1)

	results, err := redis.Values(rl.tracker.Do("EXEC"))
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
