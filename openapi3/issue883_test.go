package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	yamlv3 "go.yaml.in/yaml/v3"

	"github.com/getkin/kin-openapi/internal/yamlconv"
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

	t.Run("Roundtrip using yaml pkg", func(t *testing.T) {
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

		justPaths, err = yamlv3.Marshal(doc.Paths)
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

		marshalledYaml, err := yamlv3.Marshal(doc)
		require.NoError(t, err)
		require.NotNil(t, doc.Paths)
		require.YAMLEq(t, spec, string(marshalledYaml))

		var newDoc openapi3.T
		err = yamlv3.Unmarshal(marshalledYaml, &newDoc)
		require.NoError(t, err)
		require.NotNil(t, newDoc.Paths)
		require.Equal(t, doc, &newDoc)
	})
}
