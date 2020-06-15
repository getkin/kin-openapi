package openapi3

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var operation *Operation

func initOperation() {
	operation = NewOperation()
	operation.Description = "Some description"
	operation.Summary = "Some summary"
	operation.Tags = []string{"tag1", "tag2"}
}

func TestAddParameter(t *testing.T) {
	initOperation()
	operation.AddParameter(NewQueryParameter("param1"))
	operation.AddParameter(NewCookieParameter("param2"))
	require.Equal(t, "param1", operation.Parameters.GetByInAndName("query", "param1").Name)
	require.Equal(t, "param2", operation.Parameters.GetByInAndName("cookie", "param2").Name)
}

func TestAddResponse(t *testing.T) {
	initOperation()
	operation.AddResponse(200, NewResponse())
	operation.AddResponse(400, NewResponse())
	require.NotNil(t, "status 200", operation.Responses.Get(200).Value)
	require.NotNil(t, "status 400", operation.Responses.Get(400).Value)
}

func operationWithoutResponses() *Operation {
	initOperation()
	return operation
}

func operationWithResponses() *Operation {
	initOperation()
	operation.AddResponse(200, NewResponse().WithDescription("some response"))
	return operation
}

func TestOperationValidation(t *testing.T) {
	tests := []struct {
		name          string
		input         *Operation
		expectedError error
	}{
		{
			"when no Responses object is provided",
			operationWithoutResponses(),
			errors.New("value of responses must be a JSON object"),
		},
		{
			"when a Responses object is provided",
			operationWithResponses(),
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := context.Background()
			validationErr := test.input.Validate(c)

			require.Equal(t, test.expectedError, validationErr, "expected errors (or lack of) to match")
		})
	}
}
