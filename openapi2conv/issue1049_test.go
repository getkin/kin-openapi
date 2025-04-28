package openapi2conv

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1049(t *testing.T) {
	spec := `
{
  "swagger": "2.0",
  "info": {
    "description": "Test for additionalProperties",
    "version": "v1",
    "title": "API Test"
  },
  "host": "Test",
  "paths": {
    "/map": {
      "post": {
        "summary": "api test summary",
        "description": "api test description",
        "operationId": "apiTestUsingPOST",
        "consumes": ["application/json"],
        "produces": ["*/*"],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/ResponseOfMap"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "BaseDictResp": {
      "type": "object",
      "properties": {
        "key": {
          "type": "string",
          "description": "key of map"
        },
        "value": {
          "type": "string",
          "description": "value of map"
        }
      },
      "title": "BaseDictResp",
      "description": "BaseDictResp description"
    },

    "ResponseOfMap": {
      "type": "object",
      "properties": {
        "data": {
          "type": "object",
          "description": "response data",
          "additionalProperties": {
            "type": "array",
            "items": {
              "$ref": "#/definitions/BaseDictResp"
            }
          }
        }
      },
      "title": "ResponseOfMap"
    }
  }
}
`
	doc3, err := v2v3JSON([]byte(spec))
	require.NoError(t, err)

	spec3, err := json.Marshal(doc3)
	require.NoError(t, err)
	const expected = `{"components":{"schemas":{"BaseDictResp":{"description":"BaseDictResp description","properties":{"key":{"description":"key of map","type":"string"},"value":{"description":"value of map","type":"string"}},"title":"BaseDictResp","type":"object"},"ResponseOfMap":{"properties":{"data":{"additionalProperties":{"items":{"$ref":"#/components/schemas/BaseDictResp"},"type":"array"},"description":"response data","type":"object"}},"title":"ResponseOfMap","type":"object"}}},"info":{"description":"Test for additionalProperties","title":"API Test","version":"v1"},"openapi":"3.0.3","paths":{"/map":{"post":{"description":"api test description","operationId":"apiTestUsingPOST","responses":{"200":{"content":{"*/*":{"schema":{"$ref":"#/components/schemas/ResponseOfMap"}}},"description":"OK"}},"summary":"api test summary"}}},"servers":[{"url":"https://Test/"}]}`
	require.JSONEq(t, expected, string(spec3))

	err = doc3.Validate(context.Background())
	require.NoError(t, err)
}
