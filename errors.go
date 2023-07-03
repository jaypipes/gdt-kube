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
	// ErrInvalidKubeConfigNotFound is returned when a kubeconfig path points
	// to a file that does not exist.
	ErrInvalidKubeConfigNotFound = fmt.Errorf(
		"%w: specified kube config path not found",
		gdterrors.ErrInvalid,
	)
	// ErrInvalidResourceSpecifier is returned when the test author uses a
	// resource specifier for the `kube.get` or `kube.delete` fields that is
	// not valid.
	ErrInvalidResourceSpecifier = fmt.Errorf(
		"%w: invalid resource specifier",
		gdterrors.ErrInvalid,
	)
	// ErrInvalidResourceSpecifierOrFilepath is returned when the test author
	// uses a resource specifier for the `kube.delete` fields that is not valid
	// or is not a filepath.
	ErrInvalidResourceSpecifierOrFilepath = fmt.Errorf(
		"%w: invalid resource specifier or filepath",
		gdterrors.ErrInvalid,
	)
	// ErrResourceUnknown is returned when an unknown resource kind is
	// specified for a create/apply/delete target. This is a runtime error
	// because we rely on the discovery client to determine whether a resource
	// kind is valid.
	ErrResourceUnknown = fmt.Errorf(
		"%w: resource unknown",
		gdterrors.ErrFailure,
	)
	// ErrExpectedNotFound is returned when we expected to get either a
	// NotFound response code (get) or an empty set of results (list) but did
	// not find that.
	ErrExpectedNotFound = fmt.Errorf(
		"%w: expected not found",
		gdterrors.ErrFailure,
	)
)

// KubeConfigNotFound returns ErrInvalidKubeConfigNotFound for a given filepath
func KubeConfigNotFound(path string) error {
	return fmt.Errorf("%w: %s", ErrInvalidKubeConfigNotFound, path)
}

// InvalidResourceSpecifier returns ErrInvalidResourceSpecifier for a given
// supplied resource specifier.
func InvalidResourceSpecifier(subject string) error {
	return fmt.Errorf("%w: %s", ErrInvalidResourceSpecifier, subject)
}

// InvalidResourceSpecifierOrFilepath returns
// ErrInvalidResourceSpecifierOrFilepath for a given supplied subject.
func InvalidResourceSpecifierOrFilepath(subject string) error {
	return fmt.Errorf("%w: %s", ErrInvalidResourceSpecifierOrFilepath, subject)
}

// ResourceUnknown returns ErrRuntimeResourceUnknown for a given kind
func ResourceUnknown(kind string) error {
	return fmt.Errorf("%w: %s", ErrResourceUnknown, kind)
}

// ExpectedNotFound returns ErrExpectedNotFound for a given status code or
// number of items.
func ExpectedNotFound(msg string) error {
	return fmt.Errorf("%w: %s", ErrExpectedNotFound, msg)
}
