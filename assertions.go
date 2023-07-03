// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	gdterrors "github.com/jaypipes/gdt-core/errors"
	gdttypes "github.com/jaypipes/gdt-core/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	msgExpectedError = "Expected response to have error containing %s but got %s"
	msgExpectedLen   = "Expected response to have %d items in result but got %d"
)

// Expect contains one or more assertions about a kube client call
type Expect struct {
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

// assertions contains all assertions made for the exec test
type assertions struct {
	// failures contains the set of error messages for failed assertions
	failures []error
	// terminal indicates there was a failure in evaluating the assertions that
	// should be considered a terminal condition (and therefore the test action
	// should not be retried).
	terminal bool
	// exp contains the expected conditions to assert against
	exp *Expect
	// err is the error returned by the client or action. This is evaluated
	// against a set of expected conditions.
	err error
	// r is either an `unstructured.Unstructured` or an
	// `unstructured.UnstructuredList` response returned from the kube client
	// call.
	r interface{}
}

// Fail appends a supplied error to the set of failed assertions
func (a *assertions) Fail(err error) {
	a.failures = append(a.failures, err)
}

// Failures returns a slice of errors for all failed assertions
func (a *assertions) Failures() []error {
	if a == nil {
		return []error{}
	}
	return a.failures
}

// Terminal returns a bool indicating the assertions failed in a way that is
// not retryable.
func (a *assertions) Terminal() bool {
	if a == nil {
		return false
	}
	return a.terminal
}

// OK checks all the assertions against the supplied arguments and returns true
// if all assertions pass.
func (a *assertions) OK() bool {
	exp := a.exp
	if exp == nil {
		if a.err != nil {
			a.Fail(gdterrors.UnexpectedError(a.err))
			a.terminal = true
			return false
		}
		return true
	}
	if !a.errorOK() {
		return false
	}
	if !a.unknownOK() {
		return false
	}
	if !a.notFoundOK() {
		return false
	}
	if !a.lenOK() {
		return false
	}
	return true
}

// errorOK returns true if the supplied error matches the Error conditions,
// false otherwise.
func (a *assertions) errorOK() bool {
	exp := a.exp
	if exp.Error != "" {
		if a.err == nil {
			a.Fail(gdterrors.UnexpectedError(a.err))
			a.terminal = true
			return false
		}
		if !strings.Contains(a.err.Error(), exp.Error) {
			a.Fail(gdterrors.NotIn(a.err.Error(), exp.Error))
			return false
		}
	}
	return true
}

// unknownOK returns true if the supplied error matches the Unknown condition,
// false otherwise.
func (a *assertions) unknownOK() bool {
	exp := a.exp
	if exp.Unknown {
		if !errors.Is(a.err, ErrResourceUnknown) {
			a.Fail(ResourceUnknown(a.err.Error()))
		}
	}
	return true
}

// notFoundOK returns true if the supplied error and response matches the
// NotFound condition and the Len==0 condition, false otherwise
func (a *assertions) notFoundOK() bool {
	exp := a.exp
	if (exp.Len != nil && *exp.Len == 0) || exp.NotFound {
		// First check if the error is like one returned from Get or Delete
		// that has a 404 ErrStatus.Code in it
		apierr, ok := a.err.(*apierrors.StatusError)
		if ok {
			if http.StatusNotFound != int(apierr.ErrStatus.Code) {
				msg := fmt.Sprintf("got status code %d", apierr.ErrStatus.Code)
				a.Fail(ExpectedNotFound(msg))
				return false
			}
		}
		// Next check to see if the supplied resp is a list of objects returned
		// by the dynamic client and if so, is that an empty list.
		list, ok := a.r.(*unstructured.UnstructuredList)
		if ok {
			if len(list.Items) != 0 {
				msg := fmt.Sprintf("got %d items", len(list.Items))
				a.Fail(ExpectedNotFound(msg))
				return false
			}
		}
	}
	return true
}

// lenOK returns true if the supplied error and subject matches the Len
// condition, false otherwise
func (a *assertions) lenOK() bool {
	exp := a.exp
	if exp.Len != nil {
		// if the supplied resp is a list of objects returned by the dynamic
		// client check its length
		list, ok := a.r.(*unstructured.UnstructuredList)
		if ok {
			if len(list.Items) != *exp.Len {
				a.Fail(gdterrors.NotEqualLength(*exp.Len, len(list.Items)))
				return false
			}
		}
	}
	return true
}

// newAssertions returns an assertions object populated with the supplied http
// spec assertions
func newAssertions(
	exp *Expect,
	err error,
	r interface{},
) gdttypes.Assertions {
	return &assertions{
		failures: []error{},
		exp:      exp,
		err:      err,
		r:        r,
	}
}
