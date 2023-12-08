package openapi3_test

import (
	"testing"

	"github.com/invopop/yaml"
	"github.com/stretchr/testify/require"

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
paths: {}
components:
  schemas:
    Kitten:
      type: string
`[1:]

	sl := openapi3.NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.NotNil(t, doc.Paths)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
	require.NotNil(t, doc.Paths)

	marshalledJson, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{
  "openapi": "3.0.0",
  "info": {
    "version": "1.0.0",
    "title": "Swagger Petstore",
    "license": {"name": "MIT"}
  },
  "servers": [{"url": "http://petstore.swagger.io/v1"}],
  "paths": {},
  "components": {
    "schemas": {
      "Kitten": {"type": "string"}
    }
  }
}`, string(marshalledJson))
	require.NotNil(t, doc.Paths)

	marshalledYaml, err := yaml.Marshal(&doc)
	require.NoError(t, err)
	require.YAMLEq(t, spec, string(marshalledYaml))
	require.NotNil(t, doc.Paths)
}
