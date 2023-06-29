// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	// NotFound is a bool indicating the result of a call should be a
	// NotFound error. Alternately, the user can set `assert.len = 0` and for
	// single-object-returning calls (e.g. `get` or `delete`) the assertion is
	// equivalent to `assert.notfound = true`
	NotFound bool `yaml:"notfound,omitempty"`
	// Unknown is a bool indicating the test author expects that they will have
	// gotten an error ("the server could not find the requested resource")
	// from the Kubernetes API server. This is mostly good for unit/fuzz
	// testing CRDs.
	Unknown bool `yaml:"unknown,omitempty"`
}

func assertError(t *testing.T, exp string, got error) {
	t.Helper()
	assert.ErrorContains(t, got, exp, msgExpectedError, exp, got)
}

func assertLen(t *testing.T, exp int, got int) {
	t.Helper()
	assert.Equal(t, exp, got, msgExpectedLen, exp, got)
}

// OK returns true if all contained assertions pass successfully given the
// supplied error (returned from some function or Kubernetes client method) and
// response (either nil or a Response or an unstructured.Unstructured to
// evaluate). We return a nil error if all assertions passed, essentially
// "swallowing" the supplied error when all assertions are successful.
func (a *Assertions) OK(err error, resp interface{}) (error, bool) {
	if a == nil {
		// If we get an error and had no assertions, that's a failure.
		return err, err == nil
	}
	if !a.errorOK(err) {
		return err, false
	}
	if !a.unknownOK(err) {
		return err, false
	}
	if !a.notFoundOK(err, resp) {
		return err, false
	}
	if !a.lenOK(err, resp) {
		return err, false
	}
	return nil, true
}

// errorOK returns true if the supplied error matches the Error conditions,
// false otherwise.
func (a *Assertions) errorOK(err error) bool {
	if a == nil {
		// If we get an error and had no assertions, that's a failure.
		return err == nil
	}
	if a.Error != "" {
		if err == nil {
			// We expected an error but got none...
			return false
		}
		return strings.Contains(err.Error(), a.Error)
	}
	return true
}

// unknownOK returns true if the supplied error matches the Unknown condition,
// false otherwise.
func (a *Assertions) unknownOK(err error) bool {
	if a == nil {
		// If we get an error and had no assertions, that's a failure.
		return err == nil
	}
	if a.Unknown {
		return errors.Is(err, ErrRuntimeResourceUnknown)
	}
	return true
}

// notFoundOK returns true if the supplied error and response matches the
// NotFound condition and the Len==0 condition, false otherwise
func (a *Assertions) notFoundOK(err error, resp interface{}) bool {
	if a == nil {
		// If we get an error and had no assertions, that's a failure.
		return err == nil
	}
	if (a.Len != nil && *a.Len == 0) || a.NotFound {
		// First check if the error is like one returned from Get or Delete
		// that has a 404 ErrStatus.Code in it
		apierr, ok := err.(*apierrors.StatusError)
		if ok {
			return http.StatusNotFound == int(apierr.ErrStatus.Code)
		}
		// Next check to see if the supplied resp is a list of objects returned
		// by the dynamic client and if so, is that an empty list.
		list, ok := resp.(*unstructured.UnstructuredList)
		if ok {
			return len(list.Items) == 0
		}
	}
	return true
}

// lenOK returns true if the supplied error and subject matches the Len
// condition, false otherwise
func (a *Assertions) lenOK(err error, resp interface{}) bool {
	if a == nil {
		// If we get an error and had no assertions, that's a failure.
		return err == nil
	}
	if a.Len != nil && *a.Len > 0 {
		// First check if the error is like one returned from Get or Delete
		// that has a 404 ErrStatus.Code in it
		apierr, ok := err.(*apierrors.StatusError)
		if ok {
			if http.StatusNotFound == int(apierr.ErrStatus.Code) {
				return false
			}
		}
		// Next check to see if the supplied resp is a list of objects returned
		// by the dynamic client and if so, is that an empty list.
		list, ok := resp.(*unstructured.UnstructuredList)
		if ok {
			return len(list.Items) == *a.Len
		}
	}
	return true
}
