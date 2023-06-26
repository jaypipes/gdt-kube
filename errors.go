// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"fmt"

	gdterrors "github.com/jaypipes/gdt-core/errors"
)

var (
	// ErrInvalidMoreThanOneShortcut is returned when the test author included
	// more than one shortcut (e.g. `kube.create` or `kube.apply`) in the same
	// test spec.
	ErrInvalidMoreThanOneShortcut = fmt.Errorf(
		"%w: you may only specify a single shortcut field (e.g. "+
			"`kube.create` or `kube.apply`",
		gdterrors.ErrInvalid,
	)
	// ErrInvalidEitherShortcutOrKubeSpec is returned when the test author
	// included both a shortcut (e.g. `kube.create` or `kube.apply`) AND the
	// long-form `kube` object in the same test spec.
	ErrInvalidEitherShortcutOrKubeSpec = fmt.Errorf(
		"%w: either specify a full KubeSpec in the `kube` field or specify "+
			"one of the shortcuts (e.g. `kube.create` or `kube.apply`",
		gdterrors.ErrInvalid,
	)
	// ErrInvalidMoreThanOneKubeAction is returned when the test author
	// included more than one Kubernetes action (e.g. `create` or `apply`) in
	// the same KubeSpec.
	ErrInvalidMoreThanOneKubeAction = fmt.Errorf(
		"%w: you may only specify a single Kubernetes action field "+
			"(e.g. `create`, `apply` or `delete`) in the `kube` object. ",
		gdterrors.ErrInvalid,
	)
	ErrInvalidKubeConfigNotFound = fmt.Errorf(
		"%w: specified kube config path not found",
		gdterrors.ErrInvalid,
	)
)

// KubeConfigNotFound returns ErrInvalidKubeConfigNotFound for a given filepath
func KubeConfigNotFound(path string) error {
	return fmt.Errorf("%w: %s", ErrInvalidKubeConfigNotFound, path)
}
