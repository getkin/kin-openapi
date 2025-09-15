package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestParameterExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	param := doc.Paths.Value("/test").Post.Parameters[0].Value
	require.NotNil(t, param, "Parameter should not be nil")
	require.NotNil(t, param.Examples["Test"].Value, "Parameter example should not be nil")
	require.Equal(t, "An example", param.Examples["Test"].Value.Summary)
}

func TestParameterExampleWithContentRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	param := doc.Paths.Value("/test").Post.Parameters[1].Value
	require.NotNil(t, param, "Parameter should not be nil")
	require.NotNil(t, param.Content["application/json"].Examples["Test"].Value, "Parameter example should not be nil")
	require.Equal(t, "An example", param.Content["application/json"].Examples["Test"].Value.Summary)
}

func TestRequestBodyExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	requestBody := doc.Paths.Value("/test").Post.RequestBody.Value
	require.NotNil(t, requestBody, "Request body should not be nil")
	require.NotNil(t, requestBody.Content["application/json"].Examples["Test"].Value, "Request body example should not be nil")
	require.Equal(t, "An example", requestBody.Content["application/json"].Examples["Test"].Value.Summary)
}

func TestResponseExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	response := doc.Paths.Value("/test").Post.Responses.Value("200").Value
	require.NotNil(t, response, "Response should not be nil")
	require.NotNil(t, response.Content["application/json"].Examples["Test"].Value, "Response example should not be nil")
	require.Equal(t, "An example", response.Content["application/json"].Examples["Test"].Value.Summary)
}

func TestHeaderExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	response := doc.Paths.Value("/test").Post.Responses.Value("200").Value
	header := response.Headers["X-Test-Header"].Value
	require.NotNil(t, header, "Header should not be nil")
	require.NotNil(t, header.Examples["Test"].Value, "Header example should not be nil")
	require.Equal(t, "An example", header.Examples["Test"].Value.Summary)
}

func TestComponentExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	example := doc.Components.Examples["RefExample"].Value
	require.NotNil(t, example, "Example should not be nil")
	require.Equal(t, "An example", example.Summary)
}
