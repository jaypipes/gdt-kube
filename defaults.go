// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"github.com/jaypipes/gdt-core/errors"
	"gopkg.in/yaml.v3"
)

type kubeDefaults struct {
	// Config is the path of the kubeconfig to use in executing Kubernetes
	// client calls. If empty, typical kubeconfig path-finding is used.
	Config string `yaml:"config,omitempty"`
	// Context is the name of the kubecontext to use. If empty, the kubecontext
	// marked default in the kubeconfig is used.
	Context string `yaml:"context,omitempty"`
}

// Defaults is the known HTTP plugin defaults collection
type Defaults struct {
	kubeDefaults
}

func (d *Defaults) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.ExpectedMapAt(node)
	}
	// maps/structs are stored in a top-level Node.Content field which is a
	// concatenated slice of Node pointers in pairs of key/values.
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return errors.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "kube":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			hd := kubeDefaults{}
			if err := valNode.Decode(&hd); err != nil {
				return err
			}
			d.kubeDefaults = hd
		default:
			continue
		}
	}
	return nil
}
