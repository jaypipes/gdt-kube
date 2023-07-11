// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	gdtjson "github.com/jaypipes/gdt-core/assertion/json"
	gdterrors "github.com/jaypipes/gdt-core/errors"
	gdttypes "github.com/jaypipes/gdt-core/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	// Matches is either a string or a map[string]interface{} containing the
	// resource that the `Kube.Get` should match against. If Matches is a
	// string, the string can be either a file path to a YAML manifest or
	// inline an YAML string containing the resource fields to compare.
	//
	// Only fields present in the Matches resource are compared. There is a
	// check for existence in the retrieved resource as well as a check that
	// the value of the fields match. Only scalar fields are matched entirely.
	// In other words, you do not need to specify every field of a struct field
	// in order to compare the value of a single field in the nested struct.
	//
	// As an example, imagine you wanted to check that a Deployment resource's
	// `Status.ReadyReplicas` field was 2. You do not need to specify all other
	// `Deployment.Status` fields like `Status.Replicas` in order to match the
	// `Status.ReadyReplicas` field value. You only need to include the
	// `Status.ReadyReplicas` field in the `Matches` value as these examples
	// demonstrate:
	//
	// ```yaml
	// tests:
	//  - name: check deployment's ready replicas is 2
	//    kube:
	//      get: deployments/my-deployment
	//      assert:
	//        matches: |
	//          kind: Deployment
	//          metadata:
	//            name: my-deployment
	//          status:
	//            readyReplicas: 2
	// ```
	//
	// you don't even need to include the kind and metadata in `Matches`. If
	// missing, no kind and name matching will be performed.
	//
	// ```yaml
	// tests:
	//  - name: check deployment's ready replicas is 2
	//    kube:
	//      get: deployments/my-deployment
	//      assert:
	//        matches: |
	//          status:
	//            readyReplicas: 2
	// ```
	//
	// In fact, you don't need to use an inline multiline YAML string. You can
	// use a `map[string]interface{}` as well:
	//
	// ```yaml
	// tests:
	//  - name: check deployment's ready replicas is 2
	//    kube:
	//      get: deployments/my-deployment
	//      assert:
	//        matches:
	//          status:
	//            readyReplicas: 2
	// ```
	Matches interface{} `yaml:"matches,omitempty"`
	// JSON contains the assertions about JSON data in a response from the
	// Kubernetes API server.
	JSON *gdtjson.Expect `yaml:"json,omitempty"`
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
	if !a.lenOK() {
		return false
	}
	if !a.matchesOK() {
		return false
	}
	if !a.jsonOK() {
		return false
	}
	return true
}

// errorOK returns true if the supplied error matches the Error conditions,
// false otherwise.
func (a *assertions) errorOK() bool {
	exp := a.exp
	// We first evaluate whether an error we have received should be
	// "swallowed" because it was expected. If we still have an error after
	// swallowing all unexpected errors, then that is an unexpected error and
	// we fail.
	if a.err != nil {
		if errors.Is(a.err, ErrResourceUnknown) {
			if !exp.Unknown {
				a.Fail(a.err)
				a.terminal = true
				return false
			}
			// "Swallow" the Unknown error since we expected it.
			a.err = nil
		}
		// check if the error is like one returned from Get or Delete
		// that has a 404 ErrStatus.Code in it
		apierr, ok := a.err.(*apierrors.StatusError)
		if ok {
			if !a.expectsNotFound() {
				if http.StatusNotFound != int(apierr.ErrStatus.Code) {
					msg := fmt.Sprintf("got status code %d", apierr.ErrStatus.Code)
					a.Fail(ExpectedNotFound(msg))
					return false
				}
			}
			// "Swallow" the NotFound error since we expected it.
			a.err = nil
		}
	}
	if exp.Error != "" && a.r != nil {
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
	if a.err != nil {
		a.Fail(gdterrors.UnexpectedError(a.err))
		a.terminal = true
		return false
	}
	return true
}

func (a *assertions) expectsNotFound() bool {
	exp := a.exp
	return (exp.Len != nil && *exp.Len == 0) || exp.NotFound
}

// notFoundOK returns true if the supplied error and response matches the
// NotFound condition and the Len==0 condition, false otherwise
func (a *assertions) notFoundOK() bool {
	if a.expectsNotFound() {
		// First check if the error is like one returned from Get or Delete
		// that has a 404 ErrStatus.Code in it
		apierr, ok := a.err.(*apierrors.StatusError)
		if ok {
			if http.StatusNotFound != int(apierr.ErrStatus.Code) {
				msg := fmt.Sprintf("got status code %d", apierr.ErrStatus.Code)
				a.Fail(ExpectedNotFound(msg))
				return false
			}
			return true
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
			return true
		}
	}
	return true
}

// lenOK returns true if the subject matches the Len condition, false otherwise
func (a *assertions) lenOK() bool {
	exp := a.exp
	if exp.Len != nil {
		// if the supplied resp is a list of objects returned by the dynamic
		// client check its length
		list, ok := a.r.(*unstructured.UnstructuredList)
		if ok && list != nil {
			if len(list.Items) != *exp.Len {
				a.Fail(gdterrors.NotEqualLength(*exp.Len, len(list.Items)))
				return false
			}
		}
	}
	return true
}

// matchesOK returns true if the subject matches the Matches condition, false
// otherwise
func (a *assertions) matchesOK() bool {
	exp := a.exp
	if exp.Matches != nil && a.hasSubject() {
		matchObj := matchObjectFromAny(exp.Matches)
		res, ok := a.r.(*unstructured.Unstructured)
		if ok {
			delta := compareResourceToMatchObject(res, matchObj)
			if !delta.Empty() {
				for _, diff := range delta.Differences() {
					a.Fail(MatchesNotEqual(diff))
				}
				return false
			}
			return true
		}

		// TODO(jaypipes): if the supplied resp is a list of objects returned
		// by the dynamic client check each against the supplied matches
		// fields.
		//list, ok := a.r.(*unstructured.UnstructuredList)
		//if ok {
		//	for _, obj := range list.Items {
		//      diff := compareResourceToMatchObject(obj, matchObj)
		//
		//		a.Fail(gdterrors.NotEqualLength(*exp.Len, len(list.Items)))
		//		return false
		//	}
		//}
	}
	return true
}

// jsonOK returns true if the subject matches the JSON conditions, false
// otherwise
func (a *assertions) jsonOK() bool {
	exp := a.exp
	if exp.JSON != nil && a.hasSubject() {
		var err error
		var b []byte
		res, ok := a.r.(*unstructured.Unstructured)
		if ok {
			if b, err = json.Marshal(res); err != nil {
				panic("unable to marshal unstructured.Unstructured")
			}
		}
		ja := gdtjson.New(exp.JSON, b)
		if !ja.OK() {
			a.terminal = ja.Terminal()
			for _, f := range ja.Failures() {
				a.Fail(f)
			}
			return false
		}
	}
	return true
}

// hasSubject returns true if the assertions `r` field (which contains the
// subject of which we inspect) is not `nil`.
func (a *assertions) hasSubject() bool {
	switch a.r.(type) {
	case *unstructured.Unstructured:
		v := a.r.(*unstructured.Unstructured)
		return v != nil
	case *unstructured.UnstructuredList:
		v := a.r.(*unstructured.UnstructuredList)
		return v != nil
	}
	return false
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
