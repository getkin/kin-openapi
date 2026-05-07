package openapi3_test

import (
	"bytes"
	"os"
	"testing"

	yaml "github.com/oasdiff/yaml3"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestMarshalAdditionalProperties(t *testing.T) {
	data, err := os.ReadFile("testdata/test.openapi.additionalproperties.yml")
	require.NoError(t, err)

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	spec, err := loader.LoadFromData(data)
	require.NoError(t, err)

	err = spec.Validate(t.Context())
	require.NoError(t, err)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err = enc.Encode(spec)
	require.NoError(t, err)

	// Load the doc from the serialized yaml.
	spec2, err := loader.LoadFromData(buf.Bytes())
	require.NoError(t, err)

	err = spec2.Validate(t.Context())
	require.NoError(t, err)
}
