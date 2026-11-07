package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/getkin/kin-openapi/internal/yamlconv"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue883(t *testing.T) {
	t.Parallel()

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

	t.Run("Roundtrip through yamlconv", func(t *testing.T) {
		justPaths, err := yamlconv.Marshal(doc.Paths)
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

		marshalledYaml, err := yamlconv.Marshal(doc)
		require.NoError(t, err)
		require.NotNil(t, doc.Paths)
		require.YAMLEq(t, spec, string(marshalledYaml))

		var newDoc openapi3.T
		err = yamlconv.Unmarshal(marshalledYaml, &newDoc)
		require.NoError(t, err)
		require.NotNil(t, newDoc.Paths)
		require.Equal(t, doc, &newDoc)
	})

	t.Run("Roundtrip through the yaml pkg", func(t *testing.T) {
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

		justPaths, err = yaml.Dump(doc.Paths)
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

		marshalledYaml, err := yaml.Dump(doc)
		require.NoError(t, err)
		require.NotNil(t, doc.Paths)
		require.YAMLEq(t, spec, string(marshalledYaml))

		var newDoc openapi3.T
		err = yaml.Unmarshal(marshalledYaml, &newDoc)
		require.NoError(t, err)
		require.NotNil(t, newDoc.Paths)
		require.Equal(t, doc, &newDoc)
	})
}
