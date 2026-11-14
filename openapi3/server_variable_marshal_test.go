package openapi3_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestServerVariableMarshalDefault(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "v1"} {
		for _, metadata := range []bool{false, true} {
			t.Run(fmt.Sprintf("default=%q/metadata=%t", value, metadata), func(t *testing.T) {
				variable := &openapi3.ServerVariable{Default: openapi3.Ptr(value)}
				expected := fmt.Sprintf(`{"default":%q}`, value)
				if metadata {
					variable.Enum = []string{"", "v1"}
					variable.Description = "URL prefix"
					variable.Extensions = map[string]any{"x-prefix": true}
					expected = fmt.Sprintf(`{"default":%q,"enum":["","v1"],"description":"URL prefix","x-prefix":true}`, value)
				}
				t.Run("JSON", func(t *testing.T) {
					data, err := json.Marshal(variable)
					require.NoError(t, err)
					require.JSONEq(t, expected, string(data))
				})
				t.Run("YAML", func(t *testing.T) {
					data, err := yaml.Marshal(variable)
					require.NoError(t, err)
					require.YAMLEq(t, expected, string(data))
				})
			})
		}
	}
}

func TestServerVariableDefaultDocumentMarshal(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "v1"} {
		t.Run(fmt.Sprintf("default=%q", value), func(t *testing.T) {
			expected := fmt.Sprintf(`{"openapi":"3.0.3","info":{"title":"Prefix API","version":"1"},"paths":{"/pets":{"get":{"servers":[{"url":"https://example.com/{prefix}","variables":{"prefix":{"default":%q}}}],"responses":{"200":{"description":"OK"}}}}}}`, value)
			yamlInput := fmt.Sprintf(`openapi: 3.0.3
info: {title: Prefix API, version: '1'}
paths:
  /pets:
    get:
      servers:
      - url: https://example.com/{prefix}
        variables:
          prefix: {default: %q}
      responses:
        '200': {description: OK}
`, value)
			servers := openapi3.Servers{&openapi3.Server{
				URL:       "https://example.com/{prefix}",
				Variables: openapi3.ServerVariables{"prefix": {Default: openapi3.Ptr(value)}},
			}}
			operation := &openapi3.Operation{Servers: &servers}
			operation.AddResponse(200, openapi3.NewResponse().WithDescription("OK"))
			built := &openapi3.T{
				OpenAPI: "3.0.3",
				Info:    &openapi3.Info{Title: "Prefix API", Version: "1"},
				Paths:   openapi3.NewPaths(),
			}
			built.Paths.Set("/pets", &openapi3.PathItem{Get: operation})

			loader := openapi3.NewLoader()
			loadedJSON, err := loader.LoadFromData([]byte(expected))
			require.NoError(t, err)
			loadedYAML, err := loader.LoadFromData([]byte(yamlInput))
			require.NoError(t, err)
			for name, doc := range map[string]*openapi3.T{"built": built, "loaded JSON": loadedJSON, "loaded YAML": loadedYAML} {
				t.Run(name, func(t *testing.T) {
					require.NoError(t, doc.Validate(t.Context()))
					t.Run("JSON", func(t *testing.T) {
						data, err := json.Marshal(doc)
						require.NoError(t, err)
						require.JSONEq(t, expected, string(data))
					})
					t.Run("YAML", func(t *testing.T) {
						data, err := yaml.Marshal(doc)
						require.NoError(t, err)
						require.YAMLEq(t, expected, string(data))
					})
				})
			}
		})
	}
}

func TestServerVariableDefaultPresence(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, variable, path string
		valid                bool
		value                *string
	}{
		{name: "missing", variable: `{}`},
		{name: "empty", variable: `{"default":""}`, path: "/", valid: true, value: openapi3.Ptr("")},
		{name: "nonempty", variable: `{"default":"v1"}`, path: "/v1", valid: true, value: openapi3.Ptr("v1")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := fmt.Sprintf(`{"openapi":"3.0.3","info":{"title":"Prefix API","version":"1"},"paths":{},"servers":[{"url":"https://example.com/{prefix}","variables":{"prefix":%s}}]}`, tc.variable)
			for _, format := range []string{"JSON", "YAML"} {
				t.Run(format, func(t *testing.T) {
					data := []byte(input)
					if format == "YAML" {
						var value any
						require.NoError(t, json.Unmarshal(data, &value))
						var err error
						data, err = yaml.Marshal(value)
						require.NoError(t, err)
					}
					loader := openapi3.NewLoader()
					doc, err := loader.LoadFromData(data)
					require.NoError(t, err)
					require.Equal(t, tc.value, doc.Servers[0].Variables["prefix"].Default)
					if !tc.valid {
						require.ErrorContains(t, doc.Validate(t.Context()), "field default is required")
						return
					}
					require.NoError(t, doc.Validate(t.Context()))
					path, err := doc.Servers.BasePath()
					require.NoError(t, err)
					require.Equal(t, tc.path, path)
					encoded, err := json.Marshal(doc.Servers[0].Variables["prefix"])
					require.NoError(t, err)
					require.JSONEq(t, tc.variable, string(encoded))
				})
			}
		})
	}
}

func TestServerVariableMissingDefaultMarshal(t *testing.T) {
	t.Parallel()
	variable := &openapi3.ServerVariable{}
	require.Nil(t, variable.Default)
	require.ErrorContains(t, variable.Validate(t.Context()), "field default is required")
	data, err := json.Marshal(variable)
	require.NoError(t, err)
	require.JSONEq(t, `{}`, string(data))
	data, err = yaml.Marshal(variable)
	require.NoError(t, err)
	require.YAMLEq(t, `{}`, string(data))
}
