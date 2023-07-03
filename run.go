// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/jaypipes/gdt-core/result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Run executes the test described by the Kubernetes test. A new Kubernetes
// client request is made during this call.
func (s *Spec) Run(ctx context.Context, t *testing.T) error {
	c, err := s.connect(ctx)
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
		if s.Kube.Delete != "" {
			err = s.runDelete(ctx, t, c)
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
	c *connection,
) error {
	kind, name := splitKindName(s.Kube.Get)
	gvk := schema.GroupVersionKind{
		Kind: kind,
	}
	res, err := c.gvrFromGVK(gvk)
	a := newAssertions(s.Kube.Assert, err, nil)
	if !a.OK() {
		for _, f := range a.Failures() {
			t.Error(f)
		}
		return nil
	}
	if name == "" {
		return s.doList(ctx, t, c, res, s.Namespace())
	}
	return s.doGet(ctx, t, c, res, name, s.Namespace())
}

// doList performs the List() call and assertion check for a supplied resource
// kind and name
func (s *Spec) doList(
	ctx context.Context,
	t *testing.T,
	c *connection,
	res schema.GroupVersionResource,
	namespace string,
) error {
	list, err := c.client.Resource(res).Namespace(namespace).List(
		ctx,
		metav1.ListOptions{},
	)
	a := newAssertions(s.Kube.Assert, err, list)
	if !a.OK() {
		for _, f := range a.Failures() {
			t.Error(f)
		}
	}
	return nil
}

// doGet performs the Get() call and assertion check for a supplied resource
// kind and name
func (s *Spec) doGet(
	ctx context.Context,
	t *testing.T,
	c *connection,
	res schema.GroupVersionResource,
	name string,
	namespace string,
) error {
	obj, err := c.client.Resource(res).Namespace(namespace).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	a := newAssertions(s.Kube.Assert, err, obj)
	if !a.OK() {
		for _, f := range a.Failures() {
			t.Error(f)
		}
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
	c *connection,
) error {
	var err error
	var r io.Reader
	if probablyFilePath(s.Kube.Create) {
		path := s.Kube.Create
		f, err := os.Open(path)
		if err != nil {
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
	for _, obj := range objs {
		gvk := obj.GetObjectKind().GroupVersionKind()
		ns := obj.GetNamespace()
		if ns == "" {
			ns = s.Namespace()
		}
		res, err := c.gvrFromGVK(gvk)
		a := newAssertions(s.Kube.Assert, err, nil)
		if !a.OK() {
			for _, f := range a.Failures() {
				t.Error(f)
			}
			return nil
		}
		obj, err := c.client.Resource(res).Namespace(ns).Create(
			ctx,
			obj,
			metav1.CreateOptions{},
		)
		// TODO(jaypipes): Clearly this is applying the same assertion to each
		// object that was created, which is wrong. When I add the polymorphism
		// to the Assertions struct, I will modify this block to look for an
		// indexed set of error assertions.
		a = newAssertions(s.Kube.Assert, err, obj)
		if !a.OK() {
			for _, f := range a.Failures() {
				t.Error(f)
			}
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

// runDelete executes either Delete() call against the Kubernetes API server
// and evaluates any assertions that have been set for the returned results.
func (s *Spec) runDelete(
	ctx context.Context,
	t *testing.T,
	c *connection,
) error {
	if probablyFilePath(s.Kube.Delete) {
		path := s.Kube.Delete
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		objs, err := unstructuredFromReader(f)
		if err != nil {
			return err
		}
		for _, obj := range objs {
			gvk := obj.GetObjectKind().GroupVersionKind()
			res, err := c.gvrFromGVK(gvk)
			a := newAssertions(s.Kube.Assert, err, nil)
			if !a.OK() {
				for _, f := range a.Failures() {
					t.Error(f)
				}
				return nil
			}
			name := obj.GetName()
			ns := obj.GetNamespace()
			if ns == "" {
				ns = s.Namespace()
			}
			// TODO(jaypipes): Clearly this is applying the same assertion to each
			// object that was deleted, which is wrong. When I add the polymorphism
			// to the Assertions struct, I will modify this block to look for an
			// indexed set of error assertions.
			if err = s.doDelete(ctx, t, c, res, name, ns); err != nil {
				return err
			}
		}
		return nil
	}

	kind, name := splitKindName(s.Kube.Delete)
	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    kind,
	}
	res, err := c.gvrFromGVK(gvk)
	a := newAssertions(s.Kube.Assert, err, nil)
	if !a.OK() {
		for _, f := range a.Failures() {
			t.Error(f)
		}
		return nil
	}
	if name == "" {
		return s.doDeleteCollection(ctx, t, c, res, s.Namespace())
	}
	return s.doDelete(ctx, t, c, res, name, s.Namespace())
}

// doDelete performs the Delete() call and assertion check for a supplied
// resource kind and name
func (s *Spec) doDelete(
	ctx context.Context,
	t *testing.T,
	c *connection,
	res schema.GroupVersionResource,
	name string,
	namespace string,
) error {
	err := c.client.Resource(res).Namespace(namespace).Delete(
		ctx,
		name,
		metav1.DeleteOptions{},
	)
	a := newAssertions(s.Kube.Assert, err, nil)
	if !a.OK() {
		for _, f := range a.Failures() {
			t.Error(f)
		}
	}
	return nil
}

// doDeleteCollection performs the DeleteCollection() call and assertion check
// for a supplied resource kind
func (s *Spec) doDeleteCollection(
	ctx context.Context,
	t *testing.T,
	c *connection,
	res schema.GroupVersionResource,
	namespace string,
) error {
	err := c.client.Resource(res).Namespace(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{},
	)
	a := newAssertions(s.Kube.Assert, err, nil)
	if !a.OK() {
		for _, f := range a.Failures() {
			t.Error(f)
		}
	}
	return nil
}
