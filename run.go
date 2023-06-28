// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jaypipes/gdt-core/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
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
		if s.Kube.Create != "" {
			err = s.runCreate(ctx, t, c)
		}
	})
	return result.New(
		result.WithError(err),
	)
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
	list, err := c.Resource(res).Namespace(ns).List(
		ctx,
		metav1.ListOptions{},
	)
	assertions := s.Kube.Assert
	if assertions != nil {
		if assertions.Error != "" {
			assertError(t, assertions.Error, err)
		}
		if (assertions.Len != nil && *assertions.Len == 0) ||
			assertions.NotFound {
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
	_, err := c.Resource(res).Namespace(ns).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	assertions := s.Kube.Assert
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

// runCreate executes a Create() call against the Kubernetes API server and
// evaluates any assertions that have been set for the returned results.
func (s *Spec) runCreate(
	ctx context.Context,
	t *testing.T,
	c *dynamic.DynamicClient,
) error {
	var err error
	var r io.Reader
	if probablyFilePath(s.Kube.Create) {
		path := s.Kube.Create
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				return ManifestNotFound(path)
			}
			return err
		}
		defer f.Close()
		r = f
	} else {
		// Consider the string to be YAML/JSON content and marshal that into an
		// unstructured.Unstructured that we then pass to Create()
		r = strings.NewReader(s.Kube.Create)
	}
	objs, err := unstructuredFromReader(r)
	if err != nil {
		return err
	}
	ns := s.Namespace()
	for _, obj := range objs {
		kind := obj.GetObjectKind().GroupVersionKind().Kind
		res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: kind}
		_, err := c.Resource(res).Namespace(ns).Create(
			ctx,
			obj,
			metav1.CreateOptions{},
		)
		// TODO(jaypipe): Clearly this is applying the same assertion to each
		// object that was created, which is wrong. When I add the polymorphism
		// to the Assertions struct, I will modify this block to look for an
		// indexed set of error assertions.
		assertions := s.Kube.Assert
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
	}
	return nil
}

// unstructuredFromReader attempts to read the supplied io.Reader and unmarshal
// the content into zero or more unstructured.Unstructured objects
func unstructuredFromReader(
	r io.Reader,
) ([]*unstructured.Unstructured, error) {
	yr := yaml.NewYAMLReader(bufio.NewReader(r))

	objs := []*unstructured.Unstructured{}
	for {
		data, err := yr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		obj := &unstructured.Unstructured{}
		decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(data), len(data))
		if err = decoder.Decode(obj); err != nil {
			return nil, err
		}
		if obj.GetObjectKind().GroupVersionKind().Kind != "" {
			objs = append(objs, obj)
		}
	}

	return objs, nil
}
