// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"github.com/jaypipes/gdt-core/errors"
	gdttypes "github.com/jaypipes/gdt-core/types"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func (s *Spec) UnmarshalYAML(node *yaml.Node) error {
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
			if valNode.Kind != yaml.MappingNode {
				return errors.ExpectedMapAt(valNode)
			}
			var ks *KubeSpec
			if err := valNode.Decode(&ks); err != nil {
				return err
			}
			s.Kube = ks
		case "kube.create":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.KubeCreate = valNode.Value
		case "kube.apply":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.KubeApply = valNode.Value
		case "kube.delete":
			if valNode.Kind != yaml.ScalarNode {
				return errors.ExpectedScalarAt(valNode)
			}
			s.KubeDelete = valNode.Value
		default:
			if lo.Contains(gdttypes.BaseSpecFields, key) {
				continue
			}
			return errors.UnknownFieldAt(key, keyNode)
		}
	}
	if err := validateShortcuts(s); err != nil {
		return err
	}
	expandShortcut(s)
	if err := validateKubeSpec(s); err != nil {
		return err
	}
	return nil
}

// validateShortcuts ensures that the test author has specified only a single
// shortcut (e.g. `kube.create`) and that if a shortcut is specified, any
// long-form KubeSpec is not present.
func validateShortcuts(s *Spec) error {
	foundShortcuts := 0
	if s.KubeCreate != "" {
		foundShortcuts += 1
	}
	if s.KubeApply != "" {
		foundShortcuts += 1
	}
	if s.KubeDelete != "" {
		foundShortcuts += 1
	}
	if s.Kube == nil {
		if foundShortcuts > 1 {
			return ErrInvalidMoreThanOneShortcut
		} else if foundShortcuts == 0 {
			return ErrInvalidEitherShortcutOrKubeSpec
		}
	} else {
		if foundShortcuts > 0 {
			return ErrInvalidEitherShortcutOrKubeSpec
		}
	}
	return nil
}

// expandShortcut looks at the shortcut fields (e.g. `kube.create`) and expands
// the shortcut into a full KubeSpec.
func expandShortcut(s *Spec) {
	if s.Kube != nil {
		return
	}
	ks := &KubeSpec{}
	if s.KubeCreate != "" {
		ks.Create = s.KubeCreate
	}
	if s.KubeApply != "" {
		ks.Apply = s.KubeApply
	}
	if s.KubeDelete != "" {
		ks.Delete = s.KubeDelete
	}
	s.Kube = ks
}

// validateKubeSpec ensures that the test author has specified only a single
// action in the KubeSpec.
func validateKubeSpec(s *Spec) error {
	foundActions := 0
	if s.Kube.Create != "" {
		foundActions += 1
	}
	if s.Kube.Apply != "" {
		foundActions += 1
	}
	if s.Kube.Delete != "" {
		foundActions += 1
	}
	if foundActions > 1 {
		return ErrInvalidMoreThanOneKubeAction
	}
	return nil
}
