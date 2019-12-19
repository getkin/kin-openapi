package openapi2conv

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestConvOpenAPIV3ToV2(t *testing.T) {
	var swagger3 openapi3.Swagger
	err := json.Unmarshal([]byte(exampleV3), &swagger3)
	require.NoError(t, err)

	actualV2, err := FromV3Swagger(&swagger3)
	require.NoError(t, err)
	data, err := json.Marshal(actualV2)
	require.NoError(t, err)
	require.JSONEq(t, exampleV2, string(data))
}

func TestConvOpenAPIV2ToV3(t *testing.T) {
	var swagger2 openapi2.Swagger
	err := json.Unmarshal([]byte(exampleV2), &swagger2)
	require.NoError(t, err)

	actualV3, err := ToV3Swagger(&swagger2)
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
  "tags": [
    {
      "name": "Example",
      "description": "An example tag."
    }
  ],
  "paths": {
    "/another/{banana}/{id}": {
        "parameters": [
		  {
            "$ref": "#/parameters/banana"
          },
          {
            "in": "path",
            "name": "id",
			"type": "integer",
			"required": true
          }
		]
    },
    "/example": {
      "delete": {
        "description": "example delete",
        "responses": {
          "default": {
            "description": "default response"
          },
          "403": {
            "$ref": "#/responses/ForbiddenError"
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
        "tags": [
          "Example"
        ],
        "parameters": [
          {
            "in": "query",
            "name": "x"
          },
          {
            "in": "query",
            "name": "y",
            "description": "The y parameter",
            "type": "integer",
            "minimum": 1,
            "maximum": 10000,
            "default": 250
          },
          {
            "in": "query",
            "name": "bbox",
            "description": "Only return results that intersect the provided bounding box.",
            "maxItems": 4,
            "minItems": 4,
            "type": "array",
            "items": {
              "type": "number"
            }
          },
          {
            "in": "body",
            "name": "body",
            "schema": {}
          }
        ],
        "responses": {
          "200": {
            "description": "ok",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Item"
              }
            }
          },
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
  "responses": {
    "ForbiddenError": {
      "description": "Insufficient permission to perform the requested action.",
      "schema": {
        "$ref": "#/definitions/Error"
      }
    }
  },
  "definitions": {
    "Item": {
      "type": "object",
      "properties": {
        "foo": {
          "type": "string"
        }
      }
    },
    "Error": {
      "description": "Error response.",
      "type": "object",
      "required": [
        "message"
      ],
      "properties": {
        "message": {
          "type": "string"
        }
      }
    }
  },
  "parameters": {
    "banana": {
	  "in": "path",
      "type": "string"
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
  "openapi": "3.0.2",
  "info": {"title":"MyAPI","version":"0.1"},
  "components": {
    "responses": {
      "ForbiddenError": {
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/Error"
            }
          }
        },
        "description": "Insufficient permission to perform the requested action."
      }
    },
    "parameters": {
      "banana": {
	    "in": "path",
        "schema": {
          "type": "string"
        }
      }
    },
    "schemas": {
      "Item": {
        "type": "object",
        "properties": {
          "foo": {
            "type": "string"
          }
        }
      },
      "Error": {
        "description": "Error response.",
        "properties": {
          "message": {
            "type": "string"
          }
        },
        "required": [
          "message"
        ],
        "type": "object"
      }
    }
  },
  "tags": [
    {
      "name": "Example",
      "description": "An example tag."
    }
  ],
  "servers": [
    {
      "url": "https://test.example.com/v2"
    }
  ],
  "paths": {
    "/another/{banana}/{id}": {
        "parameters": [
		  {
            "$ref": "#/components/parameters/banana"
          },
          {
            "in": "path",
            "name": "id",
            "schema": {
              "type": "integer"
            },
			"required": true
          }
		]
    },
    "/example": {
      "delete": {
        "description": "example delete",
        "responses": {
          "default": {
            "description": "default response"
          },
          "403": {
            "$ref": "#/components/responses/ForbiddenError"
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
        "tags": [
          "Example"
        ],
        "parameters": [
          {
            "in": "query",
            "name": "x"
          },
          {
            "description": "The y parameter",
            "in": "query",
            "name": "y",
            "schema": {
              "default": 250,
              "maximum": 10000,
              "minimum": 1,
              "type": "integer"
            }
          },
          {
            "description": "Only return results that intersect the provided bounding box.",
            "in": "query",
            "name": "bbox",
            "schema": {
              "type": "array",
              "items": {
                "type": "number"
              },
              "minItems": 4,
              "maxItems": 4
            }
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
          "200": {
            "description": "ok",
            "content": {
              "application/json": {
                "schema": {
                  "items": {
                    "$ref": "#/components/schemas/Item"
                  },
                  "type": "array"
                }
              }
            }
          },
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
