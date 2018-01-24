package ratelimitMiddleware

import (
	"fmt"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

func TestRateLimiter(t *testing.T) {
	conn := redigomock.NewConn()
	mockRedisPool := redis.NewPool(func() (redis.Conn, error) {
		return conn, nil
	}, 10)

	oldNow := now
	defer func() {
		// reset to the original func when finished
		now = oldNow
	}()

	now = func() time.Time {
		// set a fixed time.Time for test
		// 2018-01-01 00:00:00 UTC
		return time.Date(2018, 1, 1, 1, 0, 0, 0, time.UTC)
	}

	duration, _ := time.ParseDuration(fmt.Sprintf("%ds", periodSeconds))

	rl := newRateLimiter("test", mockRedisPool)

	rl.Limit = 2 // decrease limit to make easier to test

	scenarios := []struct {
		description   string
		stub          func()
		reachLimit    bool
		expectedError error
	}{
		{
			description:   "it should return error when redis fails",
			reachLimit:    false,
			expectedError: fmt.Errorf("err"),
			stub: func() {
				conn.Command("MULTI").ExpectError(fmt.Errorf("err"))
			},
		},
		{
			description: "it should grant access when there's just one item on redis set",
			reachLimit:  false,
			stub: func() {
				unixNano := now().UnixNano()
				conn.Command("MULTI")
				conn.Command("ZADD", "test", unixNano, unixNano).Expect(int64(1))
				conn.Command("ZREMRANGEBYSCORE", "test", 0, unixNano-duration.Nanoseconds()).Expect(int64(1))
				conn.Command("ZRANGE", "test", 0, -1).Expect(int64(1))
				conn.Command("EXEC").ExpectSlice([]interface{}{"1"})
			},
		},

		{
			description: "it should block access when there are too many items on redis set",
			reachLimit:  true,
			stub: func() {
				unixNano := now().UnixNano()
				conn.Command("MULTI")
				conn.Command("ZADD", "test", unixNano, unixNano).Expect(int64(1))
				conn.Command("ZREMRANGEBYSCORE", "test", 0, unixNano-duration.Nanoseconds()).Expect(int64(1))
				conn.Command("ZRANGE", "test", 0, -1).Expect(int64(1))
				conn.Command("EXEC").ExpectSlice([]interface{}{"1", "2"})
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			conn.Clear()
			scenario.stub()

			err := rl.appendRequest()

			if scenario.expectedError != nil && err == nil {
				t.Errorf("expecting error")
			}

			if err != nil {
				return
			}

			if rl.limitIsReached() != scenario.reachLimit {
				t.Errorf("unexpected result; expect %t got %t", scenario.reachLimit, rl.limitIsReached())
			}
		})
	}
}
