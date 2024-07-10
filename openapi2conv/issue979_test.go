package openapi2conv

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue979(t *testing.T) {
	v2 := []byte(`
{
    "basePath": "/v2",
    "host": "test.example.com",
    "info": {
        "title": "MyAPI",
        "version": "0.1",
        "x-info": "info extension"
    },
    "paths": {
        "/foo": {
            "get": {
                "operationId": "getFoo",
                "produces": [
                    "application/pdf",
                    "application/json"
                ],
                "responses": {
                    "200": {
                        "description": "returns all information",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "default": {
                        "description": "OK"
                    }
                },
                "summary": "get foo"
            }
        }
    },
    "schemes": [
        "http"
    ],
    "swagger": "2.0"
}
`)

	var doc2 openapi2.T
	err := json.Unmarshal(v2, &doc2)
	require.NoError(t, err)

	doc3, err := ToV3(&doc2)
	require.NoError(t, err)
	err = doc3.Validate(context.Background())
	require.NoError(t, err)

	require.Equal(t, &openapi3.Types{"string"}, doc3.Paths.Value("/foo").Get.Responses.Value("200").Value.Content.Get("application/json").Schema.Value.Type)
	require.Equal(t, "binary", doc3.Paths.Value("/foo").Get.Responses.Value("200").Value.Content.Get("application/json").Schema.Value.Format)

	require.Equal(t, &openapi3.Types{"string"}, doc3.Paths.Value("/foo").Get.Responses.Value("200").Value.Content.Get("application/pdf").Schema.Value.Type)
	require.Equal(t, "binary", doc3.Paths.Value("/foo").Get.Responses.Value("200").Value.Content.Get("application/pdf").Schema.Value.Format)
}
