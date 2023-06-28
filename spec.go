// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"context"
	"path/filepath"
	"strings"

	gdtcontext "github.com/jaypipes/gdt-core/context"
	gdttypes "github.com/jaypipes/gdt-core/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeSpec is the complex type containing all of the Kubernetes-specific
// actions and assertions. Most users will use the `kube.create`, `kube.apply`
// and `kube.describe` shortcut fields.
type KubeSpec struct {
	// Namespace is a string indicating the Kubernetes namespace to use when
	// calling the Kubernetes API. If empty, any namespace specified in the
	// Defaults is used and then the string "default" is used.
	Namespace string `yaml:"namespace,omitempty"`
	// Create is a string containing a file path or raw YAML content describing
	// a Kubernetes resource to call `kubectl create` with.
	Create string `yaml:"create,omitempty"`
	// Apply is a string containing a file path or raw YAML content describing
	// a Kubernetes resource to call `kubectl apply` with.
	Apply string `yaml:"apply,omitempty"`
	// Delete is a string containing an argument to `kubectl delete` and must
	// be one of the following:
	//
	// - a file path to a manifest that will be read and the resources
	//   described in the manifest will be deleted
	// - a resource kind or kind alias, e.g. "pods", "po", followed by one of
	//   the following:
	//   * a space or `/` character followed by the resource name to delete
	//     only a resource with that name.
	//   * a space followed by `-l ` followed by a label to delete resources
	//     having such a label.
	//   * the string `--all` to delete all resources of that kind.
	Delete string `yaml:"delete,omitempty"`
	// Get is a string containing an argument to `kubectl get` and must be one
	// of the following:
	//
	// - a file path to a manifest that will be read and the resources within
	//   retrieved via `kubectl get`
	// - a resource kind or kind alias, e.g. "pods", "po", followed by one of
	//   the following:
	//   * a space or `/` character followed by the resource name to get only a
	//     resource with that name.
	//   * a space followed by `-l ` followed by a label to get resources
	//     having such a label.
	Get string `yaml:"get,omitempty"`
	// Config is the path of the kubeconfig to use in executing Kubernetes
	// client calls for this Spec. If empty, the `kube` defaults' `config`
	// value will be used. If that is empty, the following precedence is used:
	//
	// 1) KUBECONFIG environment variable pointing at a file.
	// 2) In-cluster config if running in cluster.
	// 3) $HOME/.kube/config if exists.
	Config string `yaml:"config,omitempty"`
	// Context is the name of the kubecontext to use for this Spec. If empty,
	// the `kube` defaults' `context` value will be used. If that is empty, the
	// kubecontext marked default in the kubeconfig is used.
	Context string `yaml:"context,omitempty"`
	// Assert houses the various assertions to be made about the kube client
	// call (Create, Apply, Get, etc)
	Assert *Assertions `yaml:"assert,omitempty"`
}

// Spec describes a test of a *single* Kubernetes API request and response.
type Spec struct {
	gdttypes.Spec
	// Kube is the complex type containing all of the Kubernetes-specific
	// actions and assertions. Most users will use the `kube.create`,
	// `kube.apply` and `kube.describe` shortcut fields.
	Kube *KubeSpec `yaml:"kube,omitempty"`
	// KubeCreate is a shortcut for the `KubeSpec.Create`. It can contain
	// either a file path or raw YAML content describing a Kubernetes resource
	// to call `kubectl create` with.
	KubeCreate string `yaml:"kube.create,omitempty"`
	// KubeApply is a shortcut for the `KubeSpec.Apply`. It is a string
	// containing a file path or raw YAML content describing a Kubernetes
	// resource to call `kubectl apply` with.
	KubeApply string `yaml:"kube.apply,omitempty"`
	// KubeDelete is a shortcut for the `KubeSpec.Delete`. It is a string
	// containing an argument to `kubectl delete` and must be one of the
	// following:
	//
	// - a file path to a manifest that will be read and the resources
	//   described in the manifest will be deleted
	// - a resource kind or kind alias, e.g. "pods", "po", followed by one of
	//   the following:
	//   * a space or `/` character followed by the resource name to delete
	//     only a resource with that name.
	//   * a space followed by `-l ` followed by a label to delete resources
	//     having such a label.
	//   * the string `--all` to delete all resources of that kind.
	KubeDelete string `yaml:"kube.delete,omitempty"`
}

// Title returns a good name for the Spec
func (s *Spec) Title() string {
	// If the user did not specify a name for the test spec, just default
	// it to the method and URL
	if s.Name != "" {
		return s.Name
	}
	if s.Kube == nil {
		// Shouldn't happen because of parsing, but you never know...
		return ""
	}
	if s.Kube.Create != "" {
		create := s.Kube.Create
		if probablyFilePath(create) {
			return "kube.create:" + filepath.Base(create)
		}
	}
	if s.Kube.Apply != "" {
		apply := s.Kube.Apply
		if probablyFilePath(apply) {
			return "kube.apply:" + filepath.Base(apply)
		}
	}
	if s.Kube.Delete != "" {
		delete := s.Kube.Delete
		if probablyFilePath(delete) {
			return "kube.delete:" + filepath.Base(delete)
		}
	}
	return ""
}

// poor man's quick-check of whether the action string is a file path or a YAML
// string...
func probablyFilePath(subject string) bool {
	return strings.ContainsRune(subject, '\n') || strings.ContainsRune(subject, '\r')
}

func (s *Spec) SetBase(b gdttypes.Spec) {
	s.Spec = b
}

func (s *Spec) Base() *gdttypes.Spec {
	return &s.Spec
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
