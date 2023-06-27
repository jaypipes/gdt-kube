// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube_test

import (
	"os"
	"path/filepath"
	"testing"

	gdtcontext "github.com/jaypipes/gdt-core/context"
	"github.com/jaypipes/gdt-core/scenario"
	gdtkube "github.com/jaypipes/gdt-kube"
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
	assert.NotNil(err)
	assert.ErrorContains(err, "context \"unknownctx\" does not exist")
}
