// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaypipes/gdt"
	gdtcontext "github.com/jaypipes/gdt-core/context"
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
	skipIfKind(t)
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
	require.Nil(err, "%s", err)
}

func TestGetPodNotFound(t *testing.T) {
	skipIfKind(t)
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

func TestCreateUnknownResource(t *testing.T) {
	skipIfKind(t)
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

func TestCreateUnknownResourceUsingGDT(t *testing.T) {
	// This is testing that the plugin registration for the gdt module (and
	// thus the lack of need to manually register the kube plugin) is working
	// properly.
	skipIfKind(t)
	require := require.New(t)
	kfix := kindfix.New()

	fp := filepath.Join("testdata", "create-unknown-resource.yaml")

	s, err := gdt.From(fp)
	require.Nil(err)
	require.NotNil(s)

	ctx := gdt.NewContext()
	ctx = gdt.RegisterFixture(ctx, "kind", kfix)
	err = s.Run(ctx, t)
	require.Nil(err)
}

func TestDeleteResourceNotFound(t *testing.T) {
	skipIfKind(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "delete-resource-not-found.yaml")
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

func TestDeleteUnknownResource(t *testing.T) {
	skipIfKind(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "delete-unknown-resource.yaml")
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

func TestPodCreateGetDelete(t *testing.T) {
	skipIfKind(t)
	require := require.New(t)

	fp := filepath.Join("testdata", "create-get-delete-pod.yaml")
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
	require.Nil(err, "%s", err)
}

func skipIfKind(t *testing.T) {
	_, found := os.LookupEnv("SKIP_KIND")
	if found {
		t.Skipf("skipping KinD-requiring test")
	}
}
