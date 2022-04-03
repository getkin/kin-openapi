package openapi3filter

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	"github.com/stretchr/testify/require"
)

func TestValidatingWriteRequestBodyWithWriteOnlyProperty(t *testing.T) {
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
                    "writeOnly": true
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Successfully got an account"
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

	sl := openapi3.NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.NoError(t, err)
	router, err := legacyrouter.NewRouter(doc)
	require.NoError(t, err)

	b, err := json.Marshal(Request{ID: "bt6kdc3d0cvp6u8u3ft0"})
	require.NoError(t, err)

	httpReq, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(b))
	require.NoError(t, err)
	httpReq.Header.Add(headerCT, "application/json")

	route, pathParams, err := router.FindRoute(httpReq)
	require.NoError(t, err)

	err = ValidateRequest(sl.Context, &RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
		Options:    &Options{EndpointType: WriteEndpoint},
	})
	require.NoError(t, err)

	// Try again with an insufficient length ID
	b, err = json.Marshal(Request{ID: "0cvp6u8u3ft0"})
	require.NoError(t, err)

	httpReq, err = http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(b))
	require.NoError(t, err)
	httpReq.Header.Add(headerCT, "application/json")

	err = ValidateRequest(sl.Context, &RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
		Options:    &Options{EndpointType: WriteEndpoint},
	})
	require.Error(t, err)
}

func TestValidatingReadRequestOnRequiredWriteOnlyProperty(t *testing.T) {
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
      "get": {
        "description": "Get an account",
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
                    "writeOnly": true
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Successfully got an account"
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

	sl := openapi3.NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.NoError(t, err)
	router, err := legacyrouter.NewRouter(doc)
	require.NoError(t, err)

	// Set no id because id is a required readonly field, but this is a write request
	b, err := json.Marshal(Request{ID: ""})
	require.NoError(t, err)

	httpReq, err := http.NewRequest(http.MethodGet, "/accounts", bytes.NewReader(b))
	require.NoError(t, err)
	httpReq.Header.Add(headerCT, "application/json")

	route, pathParams, err := router.FindRoute(httpReq)
	require.NoError(t, err)

	err = ValidateRequest(sl.Context, &RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
		Options:    &Options{EndpointType: ReadEndpoint},
	})
	require.NoError(t, err)
}
