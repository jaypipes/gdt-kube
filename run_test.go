// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	gdtcontext "github.com/jaypipes/gdt-core/context"
	"github.com/jaypipes/gdt-core/errors"
	"github.com/jaypipes/gdt-core/scenario"
	gdtkube "github.com/jaypipes/gdt-kube"
	kindfix "github.com/jaypipes/gdt-kube/fixtures/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnknownKubeContextInSpec(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fp := filepath.Join("testdata", "failures", "unknown-context.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(ctx, t)
	assert.NotNil(err)
	assert.ErrorContains(err, "context \"unknownctx\" does not exist")
}

func TestUnknownKubeContextInDefaults(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fp := filepath.Join("testdata", "failures", "unknown-context-in-defaults.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(ctx, t)
	require.NotNil(err)
	assert.ErrorContains(err, "context \"unknownctx\" does not exist")
}

func TestListPodsEmpty(t *testing.T) {
	skipIfNoDocker(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "list-pods-empty.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	kfix := kindfix.New()

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())
	ctx = gdtcontext.RegisterFixture(ctx, "kind", kfix)

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(ctx, t)
	require.Nil(err)
}

func TestGetPodNotFound(t *testing.T) {
	skipIfNoDocker(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "get-pod-not-found.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	kfix := kindfix.New()

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())
	ctx = gdtcontext.RegisterFixture(ctx, "kind", kfix)

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(ctx, t)
	require.Nil(err)
}

func TestCreateFileNotFound(t *testing.T) {
	skipIfNoDocker(t)
	require := require.New(t)
	assert := assert.New(t)

	fp := filepath.Join("testdata", "create-file-not-found.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	kfix := kindfix.New()

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())
	ctx = gdtcontext.RegisterFixture(ctx, "kind", kfix)

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(ctx, t)
	require.NotNil(err)
	require.IsType(err, &errors.RuntimeErrors{})
	re := err.(*errors.RuntimeErrors)
	assert.True(re.Has(gdtkube.ErrRuntimeManifestNotFound))
}

func TestCreateUnknownResource(t *testing.T) {
	skipIfNoDocker(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "create-unknown-resource.yaml")
	f, err := os.Open(fp)
	require.Nil(err)

	kfix := kindfix.New()

	ctx := gdtcontext.New()
	ctx = gdtcontext.RegisterPlugin(ctx, gdtkube.Plugin())
	ctx = gdtcontext.RegisterFixture(ctx, "kind", kfix)

	s, err := scenario.FromReader(
		f,
		scenario.WithPath(fp),
		scenario.WithContext(ctx),
	)
	require.Nil(err)
	require.NotNil(s)

	err = s.Run(ctx, t)
	require.Nil(err)
}

func skipIfNoDocker(t *testing.T) {
	_, err := exec.LookPath("docker")
	if err != nil || runtime.GOOS == "windows" {
		t.Skipf("no docker available in order to run KinD or Windows docker is hobbled")
	}
}
