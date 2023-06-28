// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"context"
	"strings"

	gdtcontext "github.com/jaypipes/gdt-core/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

// connection is a struct containing a discovery client and a dynamic client
// that the Spec uses to communicate with Kubernetes.
type connection struct {
	disco  discovery.DiscoveryInterface
	client dynamic.Interface
	// preferredVersions is a map, keyed by APIGroup, of the preferred
	// APIVersion for that group. When resources are not specified with a
	// version (or a group version), we use this as a lookup.
	preferredVersions map[string]string
}

// apiResourceFromGVK returns the metav1.APIResource (which is basically the
// GVK with the plural form of the Kind and some metadata about whether the
// resource is namespace scoped, etc) corresponding to the supplied
// GroupVersionKind. If no match could be made, returns
// ErrRuntimeResourceUnknown.
func (c *connection) apiResourceFromGVK(
	gvk schema.GroupVersionKind,
) (metav1.APIResource, error) {
	empty := metav1.APIResource{}
	var pv string
	var pvFound bool
	if gvk.Version == "" {
		pv, pvFound = c.preferredVersions[gvk.Group]
		if !pvFound {
			return empty, ResourceUnknown(gvk.Kind)
		}
		gvk.Version = pv
	}
	resources, err := c.disco.ServerResourcesForGroupVersion(
		gvk.GroupVersion().String(),
	)
	if err != nil {
		return empty, ResourceUnknown(gvk.Kind)
	}

	for _, r := range resources.APIResources {
		if strings.EqualFold(r.Kind, gvk.Kind) || strings.EqualFold(r.Name, gvk.Kind) {
			// NOTE(jaypipes): This is crazy that we need to do this, but
			// APIResource objects in the ServerResourcesForGroupVersion don't
			// necessarily have their Version fields set.
			if r.Version == "" {
				r.Version = pv
			}
			return r, nil
		}
	}

	return empty, ResourceUnknown(gvk.Kind)
}

// gvrFromGVK returns a GroupVersionResource from a GroupVersionKind, using the
// discovery client to look up the resource name (the plural of the kind). The
// returned GroupVersionResource will have the proper Group and Version filled
// in (as opposed to an APIResource which has empty Group and Version strings
// because it "inherits" its APIResourceList's GroupVersion ... ugh.)
func (c *connection) gvrFromGVK(
	gvk schema.GroupVersionKind,
) (schema.GroupVersionResource, error) {
	empty := schema.GroupVersionResource{}
	ar, err := c.apiResourceFromGVK(gvk)
	if err != nil {
		return empty, nil
	}
	gvr := schema.GroupVersionResource{
		Group:    ar.Group,
		Version:  ar.Version,
		Resource: ar.Name,
	}
	if gvr.Group == "" {
		gvr.Group = gvk.Group
	}
	if gvr.Version == "" {
		gvr.Version = gvk.Version
	}
	return gvr, nil
}

// connect returns a connection with a discovery client and a Kubernetes
// client-go DynamicClient to use in communicating with the Kubernetes API
// server configured for this Spec
func (s *Spec) connect(ctx context.Context) (*connection, error) {
	cfg, err := s.Config(ctx)
	if err != nil {
		return nil, err
	}
	c, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	disco, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	apiGroups, err := disco.ServerGroups()
	if err != nil {
		return nil, err
	}
	prefVersions := map[string]string{}
	for _, apiGroup := range apiGroups.Groups {
		prefVersions[apiGroup.Name] = apiGroup.PreferredVersion.Version
	}

	return &connection{
		disco:             disco,
		client:            c,
		preferredVersions: prefVersions,
	}, nil
}
