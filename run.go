// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"context"
	"net/http"
	"strings"
	"testing"

	gdtcontext "github.com/jaypipes/gdt-core/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

// Config returns a Kubernetes client-go rest.Config to use for this Spec. We
// evaluate where to retrieve the Kubernetes config from by looking at the
// following things, in this order:
//
// 1) The Spec.Kube.Config value
// 2) Any Fixtures that return a `kube.config` or `kube.config.bytes` state key
// 3) The Defaults.Config value
// 4) KUBECONFIG environment variable pointing at a file.
// 5) In-cluster config if running in cluster.
// 6) $HOME/.kube/config if exists.
func (s *Spec) Config(ctx context.Context) (*rest.Config, error) {
	d := fromBaseDefaults(s.Defaults)
	fixtures := gdtcontext.Fixtures(ctx)
	kctx := ""
	fixkctx := ""
	kcfgPath := ""
	fixkcfgPath := ""
	fixkcfgBytes := []byte{}

	for _, f := range fixtures {
		if f.HasState(StateKeyConfigBytes) {
			cfgBytesUntyped := f.State(StateKeyConfigBytes)
			fixkcfgBytes = cfgBytesUntyped.([]byte)
		}
		if f.HasState(StateKeyConfig) {
			cfgUntyped := f.State(StateKeyConfig)
			fixkcfgPath = cfgUntyped.(string)
		}
		if f.HasState(StateKeyContext) {
			ctxUntyped := f.State(StateKeyContext)
			fixkctx = ctxUntyped.(string)
		}
	}
	if s.Kube.Config != "" {
		kcfgPath = s.Kube.Config
	} else if fixkcfgPath != "" {
		kcfgPath = fixkcfgPath
	} else if d != nil && d.Config != "" {
		kcfgPath = d.Config
	}
	if s.Kube.Context != "" {
		kctx = s.Kube.Context
	} else if fixkctx != "" {
		kctx = fixkctx
	} else if d != nil && d.Context != "" {
		kctx = d.Context
	}
	overrides := &clientcmd.ConfigOverrides{}
	if kctx != "" {
		overrides.CurrentContext = kctx
	}
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kcfgPath != "" {
		rules.ExplicitPath = kcfgPath
	}
	if len(fixkcfgBytes) > 0 {
		cc, err := clientcmd.Load(fixkcfgBytes)
		if err != nil {
			return nil, err
		}
		return clientcmd.NewNonInteractiveClientConfig(
			*cc, "", overrides, rules,
		).ClientConfig()
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules, overrides,
	).ClientConfig()
}

// Client returns a Kubernetes client-go DynamicClient to use in communicating
// with the Kubernetes API server configured for this Spec
func (s *Spec) Client(ctx context.Context) (*dynamic.DynamicClient, error) {
	cfg, err := s.Config(ctx)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}

// Namespace returns the Kubernetes namespace to use when calling the
// Kubernetes API server. We evaluate which namespace to use by looking at the
// following things, in this order:
//
// 1) The Spec.Kube.Namespace value
// 2) The Defaults.Namespace value
// 3) Use the string "default"
func (s *Spec) Namespace() string {
	if s.Kube.Namespace != "" {
		return s.Kube.Namespace
	}
	d := fromBaseDefaults(s.Defaults)
	if d != nil && d.Namespace != "" {
		return d.Namespace
	}
	return "default"
}

// runGet executes a Get() call against the Kubernetes API server and evaluates
// any assertions that have been set for the returned results.
func (s *Spec) runGet(
	ctx context.Context,
	t *testing.T,
	c *dynamic.DynamicClient,
) error {
	ns := s.Namespace()
	kind, name := splitKindName(s.Kube.Get)
	if name == "" {
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
	} else {
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
