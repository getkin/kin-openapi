package openapi3filter

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestValidatingRequestBodyWithReadOnlyProperty(t *testing.T) {
	const spec = `{
  "openapi": "3.0.3",
  "info": {
    "version": "1.0.0",
    "title": "title",
    "description": "desc",
    "contact": {
      "email": "email"
    }
  },
  "paths": {
    "/accounts": {
      "post": {
        "description": "Create a new account",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["_id"],
                "properties": {
                  "_id": {
                    "type": "string",
                    "description": "Unique identifier for this object.",
                    "pattern": "[0-9a-v]+$",
                    "minLength": 20,
                    "maxLength": 20,
                    "readOnly": true
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Successfully created a new account"
          },
          "400": {
            "description": "The server could not understand the request due to invalid syntax",
          }
        }
      }
    }
  }
}
`

	type Request struct {
		ID string `json:"_id"`
	}

	sl := openapi3.NewSwaggerLoader()
	l, err := sl.LoadSwaggerFromData([]byte(spec))
	require.NoError(t, err)
	router := NewRouter().WithSwagger(l)

	b, err := json.Marshal(Request{ID: "bt6kdc3d0cvp6u8u3ft0"})
	require.NoError(t, err)

	httpReq, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(b))
	require.NoError(t, err)
	httpReq.Header.Add("Content-Type", "application/json")

	route, pathParams, err := router.FindRoute(httpReq.Method, httpReq.URL)
	require.NoError(t, err)

	err = ValidateRequest(sl.Context, &RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	})
	require.NoError(t, err)
}
