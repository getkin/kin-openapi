#!/bin/bash -eux
set -o pipefail

types=()
types+=("Callback")
types+=("Example")
types+=("Header")
types+=("Link")
types+=("Parameter")
types+=("RequestBody")
types+=("Response")
types+=("Schema")
types+=("SecurityScheme")

cat <<EOF
package openapi3

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
EOF

for type in "${types[@]}"; do
	cat <<EOF

func Test${type}Ref_Extensions(t *testing.T) {
	data := []byte(\`{"\$ref":"#/components/schemas/Pet","something":"integer","x-order":1}\`)

	ref := ${type}Ref{}
	err := json.Unmarshal(data, &ref)
	assert.NoError(t, err)

	// captures extension
	assert.Equal(t, "#/components/schemas/Pet", ref.Ref)
	assert.Equal(t, float64(1), ref.Extensions["x-order"])

	// does not capture non-extensions
	assert.Nil(t, ref.Extensions["something"])

	// validation
	err = ref.Validate(context.Background())
	require.EqualError(t, err, "extra sibling fields: [something]")

	err = ref.Validate(context.Background(), ProhibitExtensionsWithRef())
	require.EqualError(t, err, "extra sibling fields: [something x-order]")

	err = ref.Validate(context.Background(), AllowExtraSiblingFields("something"))
	assert.ErrorContains(t, err, "found unresolved ref") // expected since value not defined

	// non-extension not json lookable
	_, err = ref.JSONLookup("something")
	assert.Error(t, err)
EOF

	if [ "$type" != "Header" ]
	then
		cat <<EOF

	t.Run("extentions in value", func(t *testing.T) {
		ref.Value = &${type}{Extensions: map[string]interface{}{}}
		ref.Value.Extensions["x-order"] = 2.0

		// prefers the value next to the \$ref over the one in the \$ref.
		v, err := ref.JSONLookup("x-order")
		assert.NoError(t, err)
		assert.Equal(t, float64(1), v)
	})
EOF
	else
		cat <<EOF
	// Header does not have its own extensions.
EOF
	fi

	echo "}"

done
