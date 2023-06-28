// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Run executes the test described by the Kubernetes test. A new Kubernetes
// client request is made during this call.
func (s *Spec) Run(ctx context.Context, t *testing.T) error {
	var err error
	var c *dynamic.DynamicClient
	c, err = s.Client(ctx)
	if err != nil {
		return err
	}
	t.Run(s.Title(), func(t *testing.T) {
		if s.Kube.Get != "" {
			err = s.runGet(ctx, t, c)
		}
		return
	})
	return nil
}

// runGet executes either a List() or a Get() call against the Kubernetes API
// server and evaluates any assertions that have been set for the returned
// results.
func (s *Spec) runGet(
	ctx context.Context,
	t *testing.T,
	c *dynamic.DynamicClient,
) error {
	kind, name := splitKindName(s.Kube.Get)
	if name == "" {
		return s.doList(ctx, t, c, kind)
	}
	return s.doGet(ctx, t, c, kind, name)
}

// doList performs the List() call and assertion check for a supplied resource
// kind and name
func (s *Spec) doList(
	ctx context.Context,
	t *testing.T,
	c *dynamic.DynamicClient,
	kind string,
) error {
	ns := s.Namespace()
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: kind}
	assertions := s.Kube.Assert
	list, err := c.Resource(res).Namespace(ns).List(
		ctx,
		metav1.ListOptions{},
	)
	if assertions != nil {
		if assertions.Error != "" {
			assertError(t, assertions.Error, err)
		}
		if assertions.Len != nil {
			assertLen(t, *assertions.Len, len(list.Items))
		}
	} else {
		assert.Nil(t, err)
	}
	return nil
}

// doGet performs the Get() call and assertion check for a supplied resource
// kind and name
func (s *Spec) doGet(
	ctx context.Context,
	t *testing.T,
	c *dynamic.DynamicClient,
	kind string,
	name string,
) error {
	ns := s.Namespace()
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: kind}
	assertions := s.Kube.Assert
	_, err := c.Resource(res).Namespace(ns).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if assertions != nil {
		if assertions.Error != "" {
			assertError(t, assertions.Error, err)
		}
		if (assertions.Len != nil && *assertions.Len == 0) ||
			assertions.NotFound {
			apierr, ok := err.(*apierrors.StatusError)
			require.True(t, ok)
			assert.Equal(t, http.StatusNotFound, int(apierr.ErrStatus.Code))
		}
	} else {
		assert.Nil(t, err)
	}
	return nil
}

// splitKindName returns the Kind for a supplied `Get` or `Delete` command
// where the user can specify either a resource kind or alias, e.g. "pods" or
// "po", or the resource kind followed by a forward slash and a resource name.
func splitKindName(subject string) (string, string) {
	kind, name, _ := strings.Cut(subject, "/")
	return kind, name
}
