package openapi2conv

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue847(t *testing.T) {
	spec := []byte(`
{
  "swagger": "2.0",
  "info": {
    "description": "Hello Java Sec API",
    "version": "1.10",
    "title": "Swagger2 RESTful API"
  },
  "host": "localhost:8888",
  "basePath": "/",
  "paths": {
    "/UPLOAD/uploadSafe": {
      "post": {
        "operationId": "singleFileUploadSafeUsingPOST",
        "consumes": [
          "multipart/form-data"
        ],
        "parameters": [
          {
            "name": "file",
            "in": "formData",
            "description": "file",
            "required": true,
            "type": "file"
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "string"
            }
          }
        }
      }
    }
  }
}
`)
	doc3, err := v2v3JSON(spec)
	require.NoError(t, err)

	reqStr, err := json.MarshalIndent(doc3, "", "    ")
	require.NotNil(t, reqStr)
	require.NoError(t, err)

	err = doc3.Validate(context.Background())
	require.NoError(t, err)
}
