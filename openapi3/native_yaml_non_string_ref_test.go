// Copyright 2026 The kin-openapi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// A $ref whose value does not read as a string is not a reference. It must be
// rejected at load instead of being written back as one and read differently.
func TestNativeYAMLNonStringRefIsRejected(t *testing.T) {
	for _, value := range []string{"!!0 0", "0", "true", "1.5", "{}", "[]", "null"} {
		t.Run(value, func(t *testing.T) {
			doc := []byte(`openapi: 3.0.0
info:
  title: t
  version: '1'
paths: {}
components:
  schemas:
    S:
      $ref: ` + value + `
`)
			_, err := openapi3.NewLoader().LoadFromData(doc)
			require.ErrorContains(t, err, "$ref must be a string")
		})
	}
}
