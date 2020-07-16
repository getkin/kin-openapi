package openapi2conv

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func v2v3(spec2 []byte) (doc3 *openapi3.Swagger, err error) {
	var doc2 openapi2.Swagger
	if err = json.Unmarshal(spec2, &doc2); err != nil {
		return
	}
	doc3, err = ToV3Swagger(&doc2)
	return
}

func TestIssue187(t *testing.T) {
	spec := `
{
  "swagger": "2.0",
  "info": {
    "description": "Test Golang Application",
    "version": "1.0",
    "title": "Test",
    "contact": {
      "name": "Test",
      "email": "test@test.com"
    }
  },

  "paths": {
    "/me": {
        "get": {
          "description": "",
          "operationId": "someTest",
          "summary": "Some test",
          "tags": ["probe"],
          "produces": ["application/json"],
          "responses": {
            "200": {
              "description": "successful operation",
              "schema": {"$ref": "#/definitions/model.ProductSearchAttributeRequest"}
            }
          }
        }
      }
  },

  "host": "",
  "basePath": "/test",
  "definitions": {
    "model.ProductSearchAttributeRequest": {
      "type": "object",
      "properties": {
        "filterField": {
          "type": "string"
        },
        "filterKey": {
          "type": "string"
        },
        "type": {
          "type": "string"
        },
        "values": {
          "$ref": "#/definitions/model.ProductSearchAttributeValueRequest"
        }
      },
      "title": "model.ProductSearchAttributeRequest"
    },
    "model.ProductSearchAttributeValueRequest": {
      "type": "object",
      "properties": {
        "imageUrl": {
          "type": "string"
        },
        "text": {
          "type": "string"
        }
      },
      "title": "model.ProductSearchAttributeValueRequest"
    }
  }
}
`
	doc3, err := v2v3([]byte(spec))
	require.NoError(t, err)

	spec3, err := json.Marshal(doc3)
	require.NoError(t, err)
	const expected = `{"components":{"schemas":{"model.ProductSearchAttributeRequest":{"properties":{"filterField":{"type":"string"},"filterKey":{"type":"string"},"type":{"type":"string"},"values":{"$ref":"#/components/schemas/model.ProductSearchAttributeValueRequest"}},"title":"model.ProductSearchAttributeRequest","type":"object"},"model.ProductSearchAttributeValueRequest":{"properties":{"imageUrl":{"type":"string"},"text":{"type":"string"}},"title":"model.ProductSearchAttributeValueRequest","type":"object"}}},"info":{"contact":{"email":"test@test.com","name":"Test"},"description":"Test Golang Application","title":"Test","version":"1.0"},"openapi":"3.0.2","paths":{"/me":{"get":{"operationId":"someTest","responses":{"200":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/model.ProductSearchAttributeRequest"}}},"description":"successful operation"}},"summary":"Some test","tags":["probe"]}}}}`
	require.Equal(t, string(spec3), expected)

	err = doc3.Validate(context.Background())
	require.NoError(t, err)
}
