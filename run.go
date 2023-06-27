// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"context"
	"testing"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Run executes the test described by the Kubernetes test. A new Kubernetes
// client request is made during this call.
func (s *Spec) Run(ctx context.Context, t *testing.T) error {
	_, err := s.Client(ctx)
	if err != nil {
		return err
	}
	t.Run(s.Title(), func(t *testing.T) {
		return
	})
	return nil
}

// Config returns a Kubernetes client-go rest.Config to use for this Spec. We
// evaluate where to retrieve the Kubernetes config from by looking at the
// following things, in this order:
//
// 1) The Spec.Kube.Config value
// 2) The Defaults.Config value
// 3) KUBECONFIG environment variable pointing at a file.
// 4) In-cluster config if running in cluster.
// 5) $HOME/.kube/config if exists.
func (s *Spec) Config() (*rest.Config, error) {
	d := fromBaseDefaults(s.Defaults)
	kctx := ""
	if s.Kube.Context != "" {
		kctx = s.Kube.Context
	} else if d != nil && d.Context != "" {
		kctx = d.Context
	}
	kcfgPath := ""
	if s.Kube.Config != "" {
		kcfgPath = s.Kube.Config
	} else if d != nil && d.Config != "" {
		kcfgPath = d.Config
	}
	if kctx == "" {
		return clientcmd.BuildConfigFromFlags("", kcfgPath)
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kcfgPath},
		&clientcmd.ConfigOverrides{CurrentContext: kctx},
	).ClientConfig()
}

// Client returns a Kubernetes client-go DynamicClient to use in communicating
// with the Kubernetes API server configured for this Spec
func (s *Spec) Client(ctx context.Context) (*dynamic.DynamicClient, error) {
	cfg, err := s.Config()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}
