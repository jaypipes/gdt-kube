// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	gdtcontext "github.com/jaypipes/gdt-core/context"
	"github.com/jaypipes/gdt-core/errors"
	"github.com/jaypipes/gdt-core/scenario"
	"github.com/jaypipes/gdt-core/spec"
	gdttypes "github.com/jaypipes/gdt-core/types"
	gdtkube "github.com/jaypipes/gdt-kube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func currentDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func TestBadDefaults(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "bad-defaults.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.NotNil(err)
	assert.ErrorIs(err, errors.ErrInvalidExpectedMap)
	assert.Nil(s)
}

func TestFailureDefaultsConfigNotFound(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "defaults-config-not-found.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.NotNil(err)
	assert.ErrorIs(err, gdtkube.ErrInvalidKubeConfigNotFound)
	assert.ErrorIs(err, errors.ErrInvalid)
	assert.Nil(s)
}

func TestFailureBothShortcutAndKubeSpec(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "shortcut-and-long-kube.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.NotNil(err)
	assert.ErrorIs(err, gdtkube.ErrInvalidEitherShortcutOrKubeSpec)
	assert.ErrorIs(err, errors.ErrInvalid)
	assert.Nil(s)
}

func TestFailureMoreThanOneShortcut(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "more-than-one-shortcut.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.NotNil(err)
	assert.ErrorIs(err, gdtkube.ErrInvalidMoreThanOneShortcut)
	assert.ErrorIs(err, errors.ErrInvalid)
	assert.Nil(s)
}

func TestFailureMoreThanOneKubeAction(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "more-than-one-kube-action.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.NotNil(err)
	assert.ErrorIs(err, gdtkube.ErrInvalidMoreThanOneKubeAction)
	assert.ErrorIs(err, errors.ErrInvalid)
	assert.Nil(s)
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "parse.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	assert.Nil(err)
	assert.NotNil(s)

	assert.IsType(&scenario.Scenario{}, s)
	sc := s.(*scenario.Scenario)

	podYAML := `apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
   - name: nginx
     image: nginx:1.7.9
`

	expTests := []gdttypes.Spec{
		&gdtkube.Spec{
			Spec: spec.Spec{
				Index: 0,
				Name:  "create a pod from YAML using kube.create shortcut",
			},
			KubeCreate: podYAML,
			Kube: &gdtkube.KubeSpec{
				Create: podYAML,
			},
		},
		&gdtkube.Spec{
			Spec: spec.Spec{
				Index: 1,
				Name:  "apply a pod from a file using kube.apply shortcut",
			},
			KubeApply: "testdata/manifests/pod.yaml",
			Kube: &gdtkube.KubeSpec{
				Apply: "testdata/manifests/pod.yaml",
			},
		},
		&gdtkube.Spec{
			Spec: spec.Spec{
				Index: 2,
				Name:  "create a pod from YAML",
			},
			Kube: &gdtkube.KubeSpec{
				Create: podYAML,
			},
		},
		&gdtkube.Spec{
			Spec: spec.Spec{
				Index: 3,
				Name:  "apply a pod from a file",
			},
			Kube: &gdtkube.KubeSpec{
				Apply: "testdata/manifests/pod.yaml",
			},
		},
	}
	assert.Equal(expTests, sc.Tests)
}
