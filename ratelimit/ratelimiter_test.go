package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/teamwork/test"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))
}

func TestRateLimit(t *testing.T) {
	tests := []struct {
		description  string
		in           *http.Request
		getKey       func(*http.Request) string
		ignore       func(*http.Request) bool
		grantOnErr   bool
		grantFunc    func(context.Context, *Config, string, int, int) (bool, int, error)
		rates        func(req *http.Request) (int, int)
		expectedCode int
	}{
		{
			description: "it should grant access when grant func returns true",
			in:          &http.Request{RemoteAddr: "127.0.0.1"},
			ignore: func(req *http.Request) bool {
				return false
			},
			grantFunc: func(
				ctx context.Context,
				opts *Config,
				k string,
				perPeriodLocal, periodSecondsLocal int,
			) (bool, int, error) {
				return true, 0, nil
			},
			getKey: func(req *http.Request) string {
				return "test"
			},
			grantOnErr:   true,
			expectedCode: http.StatusOK,
		},
		{
			description: "it should grant access when ignore func returns true",
			in:          &http.Request{RemoteAddr: "127.0.0.1"},
			ignore: func(req *http.Request) bool {
				return true
			},
			grantFunc: func(
				ctx context.Context,
				opts *Config,
				k string,
				perPeriodLocal, periodSecondsLocal int,
			) (bool, int, error) {
				return false, 0, nil
			},
			getKey: func(req *http.Request) string {
				return "test"
			},
			grantOnErr:   true,
			expectedCode: http.StatusOK,
		},
		{
			description: "it should block access when grant func return false",
			in:          &http.Request{RemoteAddr: "127.0.0.1"},
			ignore: func(req *http.Request) bool {
				return false
			},
			grantFunc: func(
				ctx context.Context,
				opts *Config,
				k string,
				perPeriodLocal, periodSecondsLocal int,
			) (bool, int, error) {
				return false, 0, nil
			},
			getKey: func(req *http.Request) string {
				return "test"
			},
			grantOnErr:   true,
			expectedCode: http.StatusTooManyRequests,
		},
		{
			description: "it should grant access when GrantOnErr is true and redis returns err",
			in:          &http.Request{RemoteAddr: "127.0.0.1"},
			ignore: func(req *http.Request) bool {
				return false
			},
			grantFunc: func(
				ctx context.Context,
				opts *Config,
				k string,
				perPeriodLocal, periodSecondsLocal int,
			) (bool, int, error) {
				return false, 0, fmt.Errorf("test")
			},
			getKey: func(req *http.Request) string {
				return "test"
			},
			grantOnErr:   true,
			expectedCode: http.StatusOK,
		},
		{
			description: "it should block access when GrantOnErr is false and redis returns err",
			in:          &http.Request{RemoteAddr: "127.0.0.1"},
			ignore: func(req *http.Request) bool {
				return false
			},
			grantFunc: func(
				ctx context.Context,
				opts *Config,
				k string,
				perPeriodLocal, periodSecondsLocal int,
			) (bool, int, error) {
				return false, 0, fmt.Errorf("test")
			},
			getKey: func(req *http.Request) string {
				return "test"
			},
			grantOnErr:   false,
			expectedCode: http.StatusTooManyRequests,
		},
		{
			description: "it should use request rates when defined",
			in:          &http.Request{RemoteAddr: "127.0.0.1"},
			ignore: func(req *http.Request) bool {
				return false
			},
			grantFunc: func(
				ctx context.Context,
				opts *Config,
				k string,
				perPeriodLocal, periodSecondsLocal int,
			) (bool, int, error) {
				if perPeriodLocal != 2 {
					return false, 0, fmt.Errorf("unexpected perPeriod %d", perPeriodLocal)
				}

				if periodSecondsLocal != 120 {
					return false, 0, fmt.Errorf("unexpected periodSecondsLocal %d", periodSecondsLocal)
				}

				return true, 0, nil
			},
			getKey: func(req *http.Request) string {
				return "test"
			},
			grantOnErr: true,
			rates: func(req *http.Request) (perPeriod, periodSeconds int) {
				return 2, 120
			},
			expectedCode: http.StatusOK,
		},
	}

	oldGrant := grant
	defer func() {
		grant = oldGrant
	}()

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			grant = tt.grantFunc
			handler := RateLimit(Config{
				Pool:       nil,
				GetKey:     tt.getKey,
				Ignore:     tt.ignore,
				GrantOnErr: tt.grantOnErr,
				Rates:      tt.rates,
			})(handle{}).ServeHTTP

			rr := test.HTTP(t, tt.in, http.HandlerFunc(handler))
			if rr.Code != tt.expectedCode {
				t.Errorf("expected code %d, got %d", tt.expectedCode, rr.Code)
			}
		})
	}
}

func TestGrant(t *testing.T) {
	conn := redigomock.NewConn()

	mockRedisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) { return conn, nil },
	}

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

	perPeriod = 2 // decrease limit to make easier to test

	tests := []struct {
		description   string
		stub          func()
		granted       bool
		remaining     int
		expectedError error
	}{
		{
			description:   "it should return error when redis fails",
			granted:       false,
			remaining:     0,
			expectedError: fmt.Errorf("err"),
			stub: func() {
				conn.Command("MULTI").ExpectError(fmt.Errorf("err"))
			},
		},
		{
			description: "it should grant access when there's just one item on redis set",
			granted:     true,
			remaining:   1,
			stub: func() {
				unixNano := now().UnixNano()
				conn.Command("MULTI")
				conn.Command("ZADD", "test", unixNano, unixNano).Expect("QUEUED")
				conn.Command("ZREMRANGEBYSCORE", "test", 0, unixNano-duration.Nanoseconds()).Expect("QUEUED")
				conn.Command("ZRANGE", "test", 0, -1).Expect("QUEUED")
				conn.Command("EXEC").ExpectSlice(
					1, // result for zadd
					0, // result for zrem
					[]interface{}{ // result for zrange
						[]byte("1"),
					})
			},
		},

		{
			description: "it should block access when there are more elements in redis than the limit",
			granted:     false,
			remaining:   0,
			stub: func() {
				unixNano := now().UnixNano()
				conn.Command("MULTI")
				conn.Command("ZADD", "test", unixNano, unixNano).Expect(int64(1))
				conn.Command("ZREMRANGEBYSCORE", "test", 0, unixNano-duration.Nanoseconds()).Expect(int64(1))
				conn.Command("ZRANGE", "test", 0, -1).Expect(int64(1))
				conn.Command("EXEC").ExpectSlice(
					1, // result for zadd
					0, // result for zrem
					[]interface{}{ // result for zrange
						[]byte("1"), []byte("2"),
					})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			conn.Clear()
			tt.stub()

			granted, remaining, err := grant(context.Background(), &Config{Pool: mockRedisPool}, "test", 2, 60)

			if tt.expectedError != nil && !test.ErrorContains(err, tt.expectedError.Error()) {
				t.Fatalf("wrong error: %v", err)
			}

			if remaining != tt.remaining {
				t.Fatalf("unexpected remaining result; expect %d got %d", tt.remaining, remaining)
			}

			if granted != tt.granted {
				t.Errorf("unexpected granted result; expect %t got %t", tt.granted, granted)
			}
		})
	}
}
