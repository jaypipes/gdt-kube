// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	msgExpectedError = "Expected response to have error containing %s but got %s"
	msgExpectedLen   = "Expected response to have %d items in result but got %d"
)

// Assertions contains one or more assertions about a kube client call
type Assertions struct {
	// Error is a string that is expected to be returned as an error string
	// from the client call
	// TODO(jaypipes): Make this polymorphic to be either a shortcut string
	// (like this) or a struct containing individual error assertion fields.
	Error string `yaml:"error,omitempty"`
	// Len is an integer that is expected to represent the number of items in
	// the response when the Get request was translated into a List operation
	// (i.e. when the resource specified was a plural kind
	Len *int `yaml:"len,omitempty"`
	// NotFound is a bool indicating the the result of a call should be a
	// NotFound error. Alternately, the user can set `assert.len = 0` and for
	// single-object-returning calls (e.g. `get` or `delete`) the assertion is
	// equivalent to `assert.notfound = true`
	NotFound bool `yaml:"notfound,omitempty"`
}

func assertError(t *testing.T, exp string, got error) {
	t.Helper()
	assert.ErrorContains(t, got, exp, msgExpectedError, exp, got)
}

func assertLen(t *testing.T, exp int, got int) {
	t.Helper()
	assert.Equal(t, exp, got, msgExpectedLen, exp, got)
}
