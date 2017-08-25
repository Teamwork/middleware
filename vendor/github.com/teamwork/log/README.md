[![Build Status](https://travis-ci.com/Teamwork/log.svg?token=VszHEX46e27fhnkZbvFm&branch=master)](https://travis-ci.com/Teamwork/log)
[![codecov](https://codecov.io/gh/Teamwork/log/branch/master/graph/badge.svg?token=aGn2qV7lFa)](https://codecov.io/gh/Teamwork/log)

The `log` package contains Teamwork's Go logging code.

See godoc for more information, but tl;dr:

- `log.Print("foo")`, `log.Printf("%#v", "foo")` to log to Graylog
- `log.Error(errors.New("oh noes"))` and `log.Errorf(errors.New("oh noes"),
  "description")` to log to Sentry.

There's lots more though! See godoc!

Using it in your project
------------------------
1. Initialize the logger; e.g.

		err := log.Init(log.Options{
			...
		})

2. Make sure you call `log.Flush()` to ensure sure all hooks are handled on
   shutdown:

        defer log.Flush()

3. Install the echo logger When using echo; this lives in a `echologger` subpackage:

        e := echo.New()
        e.SetLogger(&echologger.EchoLogger{Entry: log.Module("echo")})
