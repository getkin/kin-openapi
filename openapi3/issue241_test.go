package openapi3_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue241(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/issue241.yml")
	require.NoError(t, err)

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	spec, err := loader.LoadFromData(data)
	require.NoError(t, err)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err = enc.Encode(spec)
	require.NoError(t, err)
	require.Equal(t, string(data), buf.String())
}
