// Package test contains various small helper functions that are useful when
// writing tests.
package test // import "github.com/teamwork/test"

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

// ErrorContains checks if the error message in got contains the text in
// expected.
//
// This is safe when got is nil. Use an empty string for expected if you want to
// test that err is nil.
func ErrorContains(got error, expected string) bool {
	if got == nil {
		return expected == ""
	}
	if expected == "" {
		return false
	}
	return strings.Contains(got.Error(), expected)
}

// Read data from a file.
func Read(t *testing.T, paths ...string) []byte {
	path := filepath.Join(paths...)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %v: %v", path, err)
	}
	return file
}
