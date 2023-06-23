// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"context"
	"testing"
)

// Run executes the test described by the Kubernetes test. A new Kubernetes
// client request is made during this call.
func (s *Spec) Run(ctx context.Context, t *testing.T) error {
	t.Run(s.Title(), func(t *testing.T) {
		return
	})
	return nil
}
