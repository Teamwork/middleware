package log

import (
	"io/ioutil"
	"time"

	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	graylog "gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

// SentryTimeout is the timeout we wait for Sentry responses
const SentryTimeout = 20 * time.Second

// Options for the logger configuration. Passed to Init.
type Options struct {
	// SentryEnabled controls if the output of Error*() calls are sent to
	// Sentry. If disabled, it will output just to stderr.
	SentryEnabled bool

	// SentryDSN is the DSN string in Sentry.
	SentryDSN string

	// SentryEnvironment is the environment in Sentry; this is just to make it
	// easier to see where an error originated.
	SentryEnvironment string

	// GraylogEnabled controls if the output of Print*() and Debug*() calls are
	// sent to Graylog. If disabled, it will output just to stdout.
	GraylogEnabled bool

	// GraylogAddress is the Graylog address to connect to. This is usually as
	// "[ip]:port" without schema.
	GraylogAddress string

	// Debug enables the logging of Debug*() calls.
	Debug bool

	// AWSRegion is added to the Sentry context to see on which region the error
	// occured. This is optional.
	AWSRegion string

	// Version is a version string to add (usually the git commit sha). This is
	// optional.
	Version string
}

// Init the logger. This will also initialize Graylog and/or Sentry connections
// if enabled.
func Init(opts Options) error {
	l := New()

	if opts.SentryEnabled {
		client, err := initSentry(opts)
		if err != nil {
			return errors.Wrap(err, "cannot initialize Sentry")
		}
		if err := l.enableSentry(client); err != nil {
			return errors.Wrap(err, "cannot enable sentry logging")
		}
	} else {
		Module("logger").Print("Sentry disabled; not logging to Sentry")
	}

	if opts.GraylogEnabled {
		if err := l.enableGraylog(opts.GraylogAddress); err != nil {
			return errors.Wrap(err, "failure logging to graylog")
		}
	} else {
		Module("logger").Print("Graylog disabled; logging only to STDOUT")
	}
	if opts.Debug {
		l.EnableDebug()
	}
	SetStandardLogger(l)

	return nil
}

// initSentry connects to the requested DSN, and sets the global sentryClient.
func initSentry(opts Options) (*raven.Client, error) {
	if opts.SentryDSN == "" {
		return nil, errors.New("No Sentry DSN provided")
	}

	client, err := raven.New(opts.SentryDSN)
	if err != nil {
		return nil, err
	}

	client.SetRelease(opts.Version)
	client.SetEnvironment(opts.SentryEnvironment)
	client.SetTagsContext(map[string]string{
		"region": opts.AWSRegion,
	})

	return client, nil
}

// enableSentry configures the logger to work with the provided Sentry DSN
func (l *Logger) enableSentry(client *raven.Client) error {
	hook, err := logrus_sentry.NewAsyncWithClientSentryHook(client, []logrus.Level{logrus.ErrorLevel})
	if err != nil {
		return err
	}
	hook.Timeout = SentryTimeout
	hook.StacktraceConfiguration = logrus_sentry.StackTraceConfiguration{
		Enable:        true,
		Level:         logrus.ErrorLevel,
		Skip:          0,
		InAppPrefixes: []string{"github.com/teamwork/desk"},
	}
	l.AddHook(hook)
	l.addFlusher(hook)
	return nil
}

// EnableDebug turns on debug output.
func (l *Logger) EnableDebug() {
	l.logrus.Level = logrus.DebugLevel
}

// enableGraylog configures the logger to log to Graylog instead of STDOUT
func (l *Logger) enableGraylog(enableGraylog string) error {
	if enableGraylog == "" {
		return errors.New("No graylog address string provided")
	}
	hook := graylog.NewAsyncGraylogHook(enableGraylog, nil)
	l.AddHook(hook)
	l.addFlusher(hook)
	l.SetOutput(ioutil.Discard)
	return nil
}

func (l *Logger) addFlusher(f flusher) {
	l.flushers = append(l.flushers, f)
}
