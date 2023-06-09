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
	assert.ErrorIs(err, errors.ErrExpectedMap)
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
	assert.ErrorIs(err, gdtkube.ErrKubeConfigNotFound)
	assert.ErrorIs(err, errors.ErrParse)
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
	assert.ErrorIs(err, gdtkube.ErrEitherShortcutOrKubeSpec)
	assert.ErrorIs(err, errors.ErrParse)
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
	assert.ErrorIs(err, gdtkube.ErrMoreThanOneShortcut)
	assert.ErrorIs(err, errors.ErrParse)
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
	assert.ErrorIs(err, gdtkube.ErrMoreThanOneKubeAction)
	assert.ErrorIs(err, errors.ErrParse)
	assert.Nil(s)
}

func TestFailureInvalidResourceSpecifierNoMultipleResources(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "invalid-resource-specifier-multiple-resources.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.NotNil(err)
	assert.ErrorIs(err, gdtkube.ErrResourceSpecifierInvalid)
	assert.ErrorIs(err, errors.ErrParse)
	require.Nil(s)
}

func TestFailureInvalidResourceSpecifierMutipleForwardSlashes(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "invalid-resource-specifier-multiple-forward-slashes.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.NotNil(err)
	assert.ErrorIs(err, gdtkube.ErrResourceSpecifierInvalid)
	assert.ErrorIs(err, errors.ErrParse)
	require.Nil(s)
}

func TestFailureInvalidDeleteNotFilepathOrResourceSpecifier(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "invalid-delete-not-filepath-or-resource-specifier.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.NotNil(err)
	assert.ErrorIs(err, gdtkube.ErrResourceSpecifierInvalidOrFilepath)
	assert.ErrorIs(err, errors.ErrParse)
	require.Nil(s)
}

func TestFailureCreateFileNotFound(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fp := filepath.Join("testdata", "failures", "create-file-not-found.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.NotNil(err)
	assert.ErrorIs(err, errors.ErrFileNotFound)
	assert.ErrorIs(err, errors.ErrParse)
	require.Nil(s)
}

func TestDeleteFileNotFound(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fp := filepath.Join("testdata", "failures", "delete-file-not-found.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.NotNil(err)
	assert.ErrorIs(err, errors.ErrFileNotFound)
	assert.ErrorIs(err, errors.ErrParse)
	require.Nil(s)
}

func TestFailureBadMatchesFileNotFound(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "bad-matches-file-not-found.yaml")
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
	assert.ErrorIs(err, errors.ErrFileNotFound)
	assert.Nil(s)
}

func TestFailureBadMatchesInvalidYAML(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "bad-matches-invalid-yaml.yaml")
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
	assert.ErrorIs(err, gdtkube.ErrMatchesInvalid)
	assert.Nil(s)
}

func TestFailureBadMatchesNotMapAny(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "failures", "bad-matches-not-map-any.yaml")
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
	assert.ErrorIs(err, gdtkube.ErrMatchesInvalid)
	assert.Nil(s)
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Setenv("pod_name", "foo")

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

	expTests := []gdttypes.TestUnit{
		&gdtkube.Spec{
			Spec: gdttypes.Spec{
				Index:    0,
				Name:     "create a pod from YAML using kube.create shortcut",
				Defaults: &gdttypes.Defaults{},
			},
			KubeCreate: podYAML,
			Kube: &gdtkube.KubeSpec{
				Create: podYAML,
			},
		},
		&gdtkube.Spec{
			Spec: gdttypes.Spec{
				Index:    1,
				Name:     "apply a pod from a file using kube.apply shortcut",
				Defaults: &gdttypes.Defaults{},
			},
			KubeApply: "testdata/manifests/nginx-pod.yaml",
			Kube: &gdtkube.KubeSpec{
				Apply: "testdata/manifests/nginx-pod.yaml",
			},
		},
		&gdtkube.Spec{
			Spec: gdttypes.Spec{
				Index:    2,
				Name:     "create a pod from YAML",
				Defaults: &gdttypes.Defaults{},
			},
			Kube: &gdtkube.KubeSpec{
				Create: podYAML,
			},
		},
		&gdtkube.Spec{
			Spec: gdttypes.Spec{
				Index:    3,
				Name:     "delete a pod from a file",
				Defaults: &gdttypes.Defaults{},
			},
			Kube: &gdtkube.KubeSpec{
				Delete: "testdata/manifests/nginx-pod.yaml",
			},
		},
		&gdtkube.Spec{
			Spec: gdttypes.Spec{
				Index:    4,
				Name:     "fetch a pod via kube.get shortcut",
				Defaults: &gdttypes.Defaults{},
			},
			KubeGet: "pods/name",
			Kube: &gdtkube.KubeSpec{
				Get: "pods/name",
			},
		},
		&gdtkube.Spec{
			Spec: gdttypes.Spec{
				Index:    5,
				Name:     "fetch a pod via long-form kube:get",
				Defaults: &gdttypes.Defaults{},
			},
			Kube: &gdtkube.KubeSpec{
				Get: "pods/name",
			},
		},
		&gdtkube.Spec{
			Spec: gdttypes.Spec{
				Index:    6,
				Name:     "fetch a pod with envvar substitution",
				Defaults: &gdttypes.Defaults{},
			},
			Kube: &gdtkube.KubeSpec{
				Get: "pods/foo",
			},
		},
	}
	assert.Equal(expTests, sc.Tests)
}
