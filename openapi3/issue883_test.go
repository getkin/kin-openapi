package openapi3_test

import (
	"testing"

	invopopYaml "github.com/invopop/yaml"
	"github.com/stretchr/testify/require"
	v3 "gopkg.in/yaml.v3"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue883(t *testing.T) {
	spec := `
openapi: '3.0.0'
info:
  version: '1.0.0'
  title: Swagger Petstore
  license:
    name: MIT
servers:
- url: http://petstore.swagger.io/v1
paths:
  /simple:
    get:
      summary: A simple GET request
      responses:
        '200':
          description: OK
`[1:]

	sl := openapi3.NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.NotNil(t, doc.Paths)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
	require.NotNil(t, doc.Paths)

	t.Run("Roundtrip invopop/yaml", func(t *testing.T) {
		justPaths, err := invopopYaml.Marshal(doc.Paths)
		require.NoError(t, err)
		require.NotNil(t, doc.Paths)
		require.YAMLEq(t, `
/simple:
  get:
    summary: A simple GET request
    responses:
      '200':
        description: OK
`[1:], string(justPaths))

		marshalledYaml, err := invopopYaml.Marshal(doc)
		require.NoError(t, err)
		require.NotNil(t, doc.Paths)
		require.YAMLEq(t, spec, string(marshalledYaml))

		var newDoc openapi3.T
		err = invopopYaml.Unmarshal(marshalledYaml, &newDoc)
		require.NoError(t, err)
		require.NotNil(t, newDoc.Paths)
		require.Equal(t, doc, &newDoc)
	})

	t.Run("Roundtrip yaml.v3", func(t *testing.T) {
		justPaths, err := doc.Paths.MarshalJSON()
		require.NoError(t, err)
		require.NotNil(t, doc.Paths)
		require.YAMLEq(t, `
/simple:
  get:
    summary: A simple GET request
    responses:
      '200':
        description: OK
`[1:], string(justPaths))

		justPaths, err = v3.Marshal(doc.Paths)
		require.NoError(t, err)
		require.NotNil(t, doc.Paths)
		require.YAMLEq(t, `
/simple:
  get:
    summary: A simple GET request
    responses:
      '200':
        description: OK
`[1:], string(justPaths))

		marshalledYaml, err := v3.Marshal(doc)
		require.NoError(t, err)
		require.NotNil(t, doc.Paths)
		require.YAMLEq(t, spec, string(marshalledYaml))

		t.Skip("TODO: impl https://pkg.go.dev/gopkg.in/yaml.v3#Unmarshaler on maplike types")
		var newDoc openapi3.T
		err = v3.Unmarshal(marshalledYaml, &newDoc)
		require.NoError(t, err)
		require.NotNil(t, newDoc.Paths)
		require.Equal(t, doc, &newDoc)
	})
}
