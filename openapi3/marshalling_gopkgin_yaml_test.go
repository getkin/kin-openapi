package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestIssue241Gopkg(t *testing.T) {
	spec := `
components:
  schemas:
    FooBar:
      properties:
        type_url:
          type: string
        value:
          format: byte
          type: string
      type: object
info:
  title: sample
  version: version not set
openapi: 3.0.3
paths: {}
`[1:]

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	yml, err := yaml.Marshal(doc)
	require.NoError(t, err)
	require.Equal(t, spec, string(yml))
}
