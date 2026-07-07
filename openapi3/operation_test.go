package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func initOperation() *openapi3.Operation {
	operation := openapi3.NewOperation()
	operation.Description = "Some description"
	operation.Summary = "Some summary"
	operation.Tags = []string{"tag1", "tag2"}
	return operation
}

func TestAddParameter(t *testing.T) {
	operation := initOperation()
	operation.AddParameter(openapi3.NewQueryParameter("param1"))
	operation.AddParameter(openapi3.NewCookieParameter("param2"))
	require.Equal(t, "param1", operation.Parameters.GetByInAndName("query", "param1").Name)
	require.Equal(t, "param2", operation.Parameters.GetByInAndName("cookie", "param2").Name)
}

func TestAddResponse(t *testing.T) {
	operation := initOperation()
	operation.AddResponse(200, openapi3.NewResponse())
	operation.AddResponse(400, openapi3.NewResponse())
	require.NotNil(t, "status 200", operation.Responses.Status(200).Value)
	require.NotNil(t, "status 400", operation.Responses.Status(400).Value)
}

func operationWithoutResponses() *openapi3.Operation {
	operation := initOperation()
	return operation
}

func operationWithResponses() *openapi3.Operation {
	operation := initOperation()
	operation.AddResponse(200, openapi3.NewResponse().WithDescription("some response"))
	return operation
}

func loadOperationDocument(t *testing.T, operationJSON string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`{
		"openapi": "3.1.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {"/test": {"get": ` + operationJSON + `}}
	}`))
	require.NoError(t, err)
	return doc
}

func TestOperationValidation(t *testing.T) {
	tests := []struct {
		name             string
		input            *openapi3.Operation
		expectedErrorMsg string // empty = expect no error
	}{
		{
			"when no Responses object is provided",
			operationWithoutResponses(),
			"value of responses must be an object",
		},
		{
			"when a Responses object is provided",
			operationWithResponses(),
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := t.Context()
			validationErr := test.input.Validate(c)

			if test.expectedErrorMsg == "" {
				require.NoError(t, validationErr)
			} else {
				require.EqualError(t, validationErr, test.expectedErrorMsg)
			}
		})
	}
}

func TestOperationResponsesValidationByOpenAPIVersion(t *testing.T) {
	newDocument := func(version string, responses *openapi3.Responses) *openapi3.T {
		return &openapi3.T{
			OpenAPI: version,
			Info:    &openapi3.Info{Title: "Test", Version: "1.0.0"},
			Paths: openapi3.NewPaths(openapi3.WithPath("/test", &openapi3.PathItem{
				Get: &openapi3.Operation{Responses: responses},
			})),
		}
	}

	t.Run("OpenAPI 3.0 requires responses", func(t *testing.T) {
		err := newDocument("3.0.3", nil).Validate(t.Context())
		var target *openapi3.OperationResponsesRequired
		require.ErrorAs(t, err, &target)
	})

	t.Run("OpenAPI 3.1 allows responses to be omitted", func(t *testing.T) {
		err := newDocument("3.1.0", nil).Validate(t.Context())
		require.NoError(t, err)
	})

	t.Run("OpenAPI 3.2 allows responses to be omitted", func(t *testing.T) {
		err := newDocument("3.2.0", nil).Validate(t.Context())
		require.NoError(t, err)
	})

	t.Run("OpenAPI 3.1 rejects an empty responses object", func(t *testing.T) {
		err := newDocument("3.1.0", openapi3.NewResponses()).Validate(t.Context())
		var target *openapi3.ResponsesNonEmptyRequired
		require.ErrorAs(t, err, &target)
	})
}

func TestOperationResponsesParsedStates(t *testing.T) {
	t.Run("omitted", func(t *testing.T) {
		doc := loadOperationDocument(t, `{}`)
		require.NoError(t, doc.Validate(t.Context()))
	})

	t.Run("null", func(t *testing.T) {
		doc := loadOperationDocument(t, `{"responses": null}`)
		err := doc.Validate(t.Context())
		var target *openapi3.OperationResponsesRequired
		require.ErrorAs(t, err, &target)
	})

	t.Run("empty object", func(t *testing.T) {
		doc := loadOperationDocument(t, `{"responses": {}}`)
		err := doc.Validate(t.Context())
		var target *openapi3.ResponsesNonEmptyRequired
		require.ErrorAs(t, err, &target)
	})

	t.Run("non-empty object", func(t *testing.T) {
		doc := loadOperationDocument(t, `{"responses": {"200": {"description": "ok"}}}`)
		require.NoError(t, doc.Validate(t.Context()))
	})
}

func TestOperationResponsesCanBeOmittedAfterParsing(t *testing.T) {
	for _, operationJSON := range []string{
		`{"responses": {"200": {"description": "ok"}}}`,
		`{"responses": null}`,
	} {
		doc := loadOperationDocument(t, operationJSON)
		doc.Paths.Value("/test").Get.Responses = nil
		require.NoError(t, doc.Validate(t.Context()))
	}
}

func TestOperationMarshalResponses(t *testing.T) {
	t.Run("omitted", func(t *testing.T) {
		data, err := json.Marshal(openapi3.NewOperation())
		require.NoError(t, err)
		require.JSONEq(t, `{}`, string(data))
	})

	t.Run("null", func(t *testing.T) {
		doc := loadOperationDocument(t, `{"responses": null}`)
		operation := doc.Paths.Value("/test").Get
		data, err := json.Marshal(operation)
		require.NoError(t, err)
		require.JSONEq(t, `{"responses": null}`, string(data))

		data, err = json.Marshal(operation.Responses)
		require.NoError(t, err)
		require.JSONEq(t, `null`, string(data))
	})
}

func TestOperationResponsesCanBeRepairedAfterParsingNull(t *testing.T) {
	doc := loadOperationDocument(t, `{"responses": null}`)
	operation := doc.Paths.Value("/test").Get
	operation.AddResponse(200, openapi3.NewResponse().WithDescription("ok"))
	require.NoError(t, doc.Validate(t.Context()))

	data, err := json.Marshal(operation)
	require.NoError(t, err)
	require.JSONEq(t, `{"responses":{"200":{"description":"ok"}}}`, string(data))
}

func TestOperationMissingResponsesWithMultiError(t *testing.T) {
	err := openapi3.NewOperation().Validate(t.Context(), openapi3.EnableMultiError())
	var target *openapi3.OperationResponsesRequired
	require.ErrorAs(t, err, &target)
}
