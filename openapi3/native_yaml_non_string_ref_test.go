// Copyright 2026 The kin-openapi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A $ref whose value does not read as a string is not a reference. The yaml
// path reads such a value as its text, so keeping the key would write back a
// reference that the reader takes for an external one, and the document would
// no longer read the way it was written.
func TestNativeYAMLNonStringRef(t *testing.T) {
	doc := []byte(`openapi: 3.0.0
info:
  title: t
  version: '1'
paths: {}
components:
  schemas:
    S:
      $ref: !!0 0
`)
	loader := NewLoader()
	got, err := loader.LoadFromData(doc)
	require.NoError(t, err)

	once, err := got.MarshalJSON()
	require.NoError(t, err)
	require.NotContains(t, string(once), `"$ref":"0"`)

	// The document reads back the way it was written.
	strict := NewLoader()
	_, err = strict.LoadFromData(once)
	require.NoError(t, err)
}
