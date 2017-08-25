// Package log is a customized wrapper around github.com/sirupsen/logrus
/*
General Information

This package supports three log levels: Debug, Info (accessed via the Print methods),
and Error. Each of these has three calling variants. Examples:

    log.Print("A log message for foo")
    log.Printf("A log message for %s", moduleName)
    log.Println("A log message for", moduleName)

Implementation & Reason

This module is mostly a wrapper around the github.com/sirupsen/logrus package.
The primary difference is that this module will extract stack traces from errors
produced by github.com/pkg/errors, or will add stack traces to other errors.
Other differences include:

- Teamwork-centric configuration
- It only exposes 3 log levels: Debug, Error, and Print
- It provides a custom log formatter, used when logging to stderr, for a log
  format we prefer.
- It automatically includes caller information for all Debug messages logged.
- It provides a ErrorWithFields() wrapper, which encapsulates logging keys with
  an error, to be (possibly) logged later.

Debug Logs

The log.Debug*() methods are only enabled when Debug mode is configured (by default
in development). Output includes the top frame of a stack trace, for easier debugging.

Info Logs

The log.Print*() methods by default send to standard output (and thus greylog, in
production).

Error Logs

The log.Error*() methods take a required first artument of an error type, followed
by the standard message/formatting arguments. In production, these are sent to
standard output, and to Sentry. These should only be used for actual errors that
the development team should notice and respond to. Errors that are recovered, such
as invalid UTF-8 sequences, should be logged with log.Print*, if at all.

All error logs also include a full stack trace of the error, for easier debugging.
Read more about how to interpret these stack traces below.

There is one additional error method, log.Err(), which takes only an error, and
no message or formatting variables. In this case, err.Error() is used as the
message. This should be used when the error message received already contains
sufficient context, and adding a custom message would simply be redundant. In
most cases, log.Error*() should be used instead.

In the case that an error is logged immediately, one might use the following
recipe:

    func foo(count int) {
        if count < 0 {
            log.Err(errors.New("Count is invalid"))
        }
        // Do something
    }

Simple Usage

For most common cases, you can simply call one of the log.Debug*(), log.Print*(),
or log.Error*() methods:

    func foo() {
        log.Debug("Attempting to foo")
        err, count := doSomething()
        if err != nil {
            log.Error(err, "Error doing something")
            return
        }
        log.Printf("Successfully did %d things", count)
    }

Advanced Usage

In addition to these simple uses cases, logrus affords us the flexibility to include
arbitrary information in logs. This is done with the WithField() and WithFields()
methods. Please use these options, especially for error, to make debugging easier.

To include an additional field, simply insert WithField(key, value) in your log
call chain.  Example:

    if err := parseString(str); err != nil {
        log.WithField("string", str).Error(err, "Could not parse the string")
    }

WithFields() works exactly the same, but allows assigning multiple fields at once:

    if err := sendEmail(to, from, subject); err != nil {
        log.WithFields(log.Fields{
            "to":      to,
            "from":    from,
            "subject": subject,
        }).Error(err, "Unable to send email")
    }

Two shortcuts exist for WithField:  Module() and WithError(), which set the
"module", and "error" keys respetively.

It is rare that you will want to call WithError(), as Err(), Error(), Errorf(),
and Errorln() handle this for you.

Module() is intended for use when you want to register which module generated a
log, perhaps for easier searching of related logs.  In the following example,
a log value is initialized with an included Module, then multiple logs are
created, which will contain the same Module key/value.

    func foo() {
        log := log.Module("foo")
        log.Print("Starting")
        // Do stuff
        log.Print("Done")
    }

Chaining

Because these methods are all chainable, it is also possible to pass logging
context around in a logging object, if desired.  For instance, you may wish to
set context at the beginning of an operation, and use it throughout:

    func foo(id int, ctx interface{}) {
        l := log.WithFields(log.Fields{
            "id":       id,
            "context":, ctx,
        })
        l.Printf("Starting")
        defer l.Printf("Finshed")
        if err := subFoo(l, ctx); err != nil {
            l.Errorf(err, "Problem subFooing")
        }
    }

    func subFoo(l *log.Entry, ctx interface{}) {
        l = l.Module("subFoo")
        l.Print("Starting")
        defer l.Printf("Finishing")
        // Do stuff
    }

Stack Traces & Causing Errors

Whenever an error is logged, this module will examine look for a stack trace
within the error value. Therefore, you are advised to use the
github.com/pkg/errors package at all times, when creating and returning errors,
as this package includes stack traces within error values.

If a stack trace is found within the error, the earliest stack trace available
will be included in the log message. If no stack trace is found, one will be
generated, relative to the log.Error*() call itself.

Also, as a benefit of using github.com/pkg/errors to wrap errors, it is possible
to find the original cause of an error, and when available, this will be reported
in the log as the 'cause' field.

Therefore, you are requested to make yourself familiar with this package
(Documentation here: https://godoc.org/github.com/pkg/errors), and use it at all
times in place of the standard library 'errors' package, and in place of
'fmt.Errorf'. When creating custom error types, it is advisable to honor the
causer interface
(https://godoc.org/github.com/pkg/errors#hdr-Retrieving_the_cause_of_an_error)
and the stackTracer interface
(https://godoc.org/github.com/pkg/errors#hdr-Retrieving_the_stack_trace_of_an_error_or_wrapper),
for the greatest compatibility with this logging module.

Error/Field bundling

You can bundle additional context to be logged, along with an error, using the
`ErrorWithFields()` method provided by this package:

    func foo(ctx *Context) error {
        if err := bar(); err != nil {
            return log.ErrorWithFields(err, log.Fields{
                "context": ctx,
            })
        }
    }

By doing this, if/whenever `log.WithError()`, or `log.Error()` (or a variant)
is called, any fields bundled with the error will be passed to `log.WithFields`.
This happens in bottom-up order, meaning that if an error is bundled with fields
more than once, the most recently bundled key takes precident.  Example:

    err := errors.New("sample error")
    err = log.ErrorWithFields(err, log.Fields{
        "foo": 10,
        "bar": 100,
    }) // Field to be logged: foo=10, bar=100
    err = log.ErrorWithFields(err, log.Fields{
        "foo": 20,
        "baz": 900
    }) // Fields to be logged = foo=20, bar=100, baz=9000

You can also do this from a *Entry, to retain any set fields. This makes it
possible to do something like:

    func foo(ctx *Context) error {
        log := log.Module("foo")
            if err := bar(); err != nil {
                return log.ErrorWithFields(err, log.Fields{
                    "context": ctx,
            })
        }
    }

In the above example, the returned error will contain both the "context" and
"module" fields.

Third Party Modules

This module can also produce standard *log.Logger structs, for use by third-party
modules.  Example:

    l := log.Module("database")
    logger, closeLogger := l.InfoLogger() // logger is a *log.Logger instance
                                          // which logs to InfoLevel.
    defer closeLogger()
    db.SetLogger(logger)
    // ... Do stuff which may cuase db to log

Note that it is important to call the closeLogger() instance when you are done
with it, to free the underlying io.Writer.

DebugLogger(), InfoLogger() and ErrorLogger() are all available for this
purpose, and each logs to its respective log level.

Note, however, that stack traces are not generated in these cases, and underlying
errors are not inspected for causes. This is becuase when called in this way,
all that is available to this logging module is the text to be logged. Error
objects, formatting strings, etc, are all obscured by the standard Go library's
log module.  (Which makes me wish that log.Logger was an interface. What were
those guys thinking?)

Using it

The log module can be re-used in other projects:

1. Initialize the logger; e.g.

      must.Must(func() error {
		  return log.Init(log.Options{
			  ...
		  })
	  })

2. Make sure you call `log.Flush()` to make sure all hooks are handled on
   shutdown:

       defer log.Flush()

3. Install the echo logger When using echo; this lives in a `echologger` subpackage:

       e := echo.New()
       e.SetLogger(&echologger.EchoLogger{Entry: log.Module("echo")})
*/
package log
