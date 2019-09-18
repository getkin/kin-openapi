package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestConvOpenAPIV3ToV2(t *testing.T) {
	var swagger3 openapi3.Swagger
	err := json.Unmarshal([]byte(exampleV3), &swagger3)
	require.NoError(t, err)

	actualV2, err := openapi2conv.FromV3Swagger(&swagger3)
	require.NoError(t, err)
	data, err := json.Marshal(actualV2)
	require.NoError(t, err)
	require.JSONEq(t, exampleV2, string(data))
}

func TestConvOpenAPIV2ToV3(t *testing.T) {
	var swagger2 openapi2.Swagger
	err := json.Unmarshal([]byte(exampleV2), &swagger2)
	require.NoError(t, err)

	actualV3, err := openapi2conv.ToV3Swagger(&swagger2)
	require.NoError(t, err)
	data, err := json.Marshal(actualV3)
	require.NoError(t, err)
	require.JSONEq(t, exampleV3, string(data))
}

const exampleV2 = `
{
  "info": {"title":"MyAPI","version":"0.1"},
  "schemes": ["https"],
  "host": "test.example.com",
  "basePath": "/v2",
  "paths": {
    "/example": {
      "delete": {
        "description": "example delete",
        "responses": {
          "default": {
            "description": "default response"
          },
          "404": {
            "description": "404 response"
          }
        }
      },
      "get": {
        "operationId": "example-get",
        "summary": "example get",
        "description": "example get",
        "parameters": [
          {
            "in": "query",
            "name": "x"
          },
          {
            "in": "body",
            "name": "body",
            "schema": {}
          }
        ],
        "responses": {
          "default": {
            "description": "default response"
          },
          "404": {
            "description": "404 response"
          }
        },
        "security": [
          {
            "get_security_0": [
              "scope0",
              "scope1"
            ],
            "get_security_1": []
          }
        ]
      },
      "head": {
        "description": "example head",
        "responses": {}
      },
      "patch": {
        "description": "example patch",
        "responses": {}
      },
      "post": {
        "description": "example post",
        "responses": {}
      },
      "put": {
        "description": "example put",
        "responses": {}
      },
      "options": {
        "description": "example options",
        "responses": {}
      }
    }
  },
  "security": [
    {
      "default_security_0": [
        "scope0",
        "scope1"
      ],
      "default_security_1": []
    }
  ]
}
`

const exampleV3 = `
{
  "openapi": "3.0",
  "info": {"title":"MyAPI","version":"0.1"},
  "components": {},
  "servers": [
    {
      "url": "https://test.example.com/v2"
    }
  ],
  "paths": {
    "/example": {
      "delete": {
        "description": "example delete",
        "responses": {
          "default": {
            "description": "default response"
          },
          "404": {
            "description": "404 response"
          }
        }
      },
      "get": {
        "operationId": "example-get",
        "summary": "example get",
        "description": "example get",
        "parameters": [
          {
            "in": "query",
            "name": "x"
          }
        ],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {}
            }
          }
        },
        "responses": {
          "default": {
            "description": "default response"
          },
          "404": {
            "description": "404 response"
          }
        },
        "security": [
          {
            "get_security_0": [
              "scope0",
              "scope1"
            ],
            "get_security_1": []
          }
        ]
      },
      "head": {
        "description": "example head",
        "responses": {}
      },
      "options": {
        "description": "example options",
        "responses": {}
      },
      "patch": {
        "description": "example patch",
        "responses": {}
      },
      "post": {
        "description": "example post",
        "responses": {}
      },
      "put": {
        "description": "example put",
        "responses": {}
      }
    }
  },
  "security": [
    {
      "default_security_0": [
        "scope0",
        "scope1"
      ],
      "default_security_1": []
    }
  ]
}
`
