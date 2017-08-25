// Package testhook is a copy of github.com/sirupsen/logrus/hooks/test, modified
// to be race-safe.
package testhook

import (
	"io/ioutil"
	"sync"

	"github.com/sirupsen/logrus"
)

// Hook is a hook designed for dealing with logs in test scenarios.
type Hook struct {
	entries []*logrus.Entry
	mu      sync.RWMutex
}

// NewGlobal installs a test hook for the global logger.
func NewGlobal() *Hook {
	hook := new(Hook)
	logrus.AddHook(hook)

	return hook
}

// NewLocal installs a test hook for a given local logger.
func NewLocal(logger *logrus.Logger) *Hook {

	hook := new(Hook)
	logger.Hooks.Add(hook)

	return hook

}

// NewNullLogger creates a discarding logger and installs the test hook.
func NewNullLogger() (*logrus.Logger, *Hook) {

	logger := logrus.New()
	logger.Out = ioutil.Discard

	return logger, NewLocal(logger)

}

// Fire logs the entry.
func (t *Hook) Fire(e *logrus.Entry) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	entryCopy := *e // For concurrency safety, since logrus modifies e without any protection
	t.entries = append(t.entries, &entryCopy)
	return nil
}

// Levels returns the levels configured by this hook (which is all of them).
func (t *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// LastEntry returns the last entry that was logged or nil.
func (t *Hook) LastEntry() *logrus.Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	i := len(t.entries) - 1
	if i < 0 {
		return nil
	}
	e := *t.entries[i]
	return &e
}

// AllEntries returns all entries that were logged.
func (t *Hook) AllEntries() []*logrus.Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	// Make a copy so the returned value won't race with future log requests
	entries := make([]*logrus.Entry, len(t.entries))
	for i, entry := range t.entries {
		// Make a copy, for safety
		e := *entry
		entries[i] = &e
	}
	return entries
}

// Reset removes all Entries from this test hook.
func (t *Hook) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = make([]*logrus.Entry, 0)
}
