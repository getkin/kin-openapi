package openapi3_test

import (
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
