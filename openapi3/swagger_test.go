package openapi3_test

import (
	"encoding/json"
	"github.com/jban332/kincore/jsontest"
	"github.com/jban332/kinapi/jsoninfo"
	"github.com/jban332/kinapi/openapi3"
	"testing"
)

type Array []interface{}
type Object map[string]interface{}

func expect(t *testing.T, swagger *openapi3.Swagger, value interface{}) {
	t.Log("Marshal *openapi3.Swagger to JSON")
	data, err := json.Marshal(swagger)
	jsontest.ExpectWithErr(t, json.RawMessage(data), err).Value(value)

	t.Log("Unmarshal *openapi3.Swagger from JSON")
	swagger = &openapi3.Swagger{}
	err = json.Unmarshal(data, &swagger)
	jsontest.ExpectWithErr(t, swagger, err).Value(value)
}

func TestRefs(t *testing.T) {
	parameter := &openapi3.Parameter{
		RefProps: jsoninfo.RefProps{
			Ref: "#/components/parameters/someParameter",
		},
		Description: "Some parameter",
	}
	requestBody := &openapi3.RequestBody{
		RefProps: jsoninfo.RefProps{
			Ref: "#/components/requestBodies/someRequestBody",
		},
		Description: "Some request body",
	}
	response := &openapi3.Response{
		RefProps: jsoninfo.RefProps{
			Ref: "#/components/responses/someResponse",
		},
		Description: "Some response",
	}
	schema := &openapi3.Schema{
		RefProps: jsoninfo.RefProps{
			Ref: "#/components/schemas/someSchema",
		},
		Description: "Some schema",
	}
	swagger := &openapi3.Swagger{
		OpenAPI: "3.0",
		Paths: openapi3.Paths{
			"/hello": &openapi3.PathItem{
				Get: &openapi3.Operation{
					Parameters: openapi3.Parameters{
						parameter,
					},
				},
			},
		},
		Components: openapi3.Components{
			Parameters: map[string]*openapi3.Parameter{
				"someParameter": parameter,
			},
			RequestBodies: map[string]*openapi3.RequestBody{
				"someRequestBody": requestBody,
			},
			Responses: map[string]*openapi3.Response{
				"someResponse": response,
			},
			Schemas: map[string]*openapi3.Schema{
				"someSchema": schema,
			},
		},
	}
	expect(t, swagger, Object{
		"openapi": "3.0",
		"info":    Object{},
		"paths": Object{
			"/hello": Object{
				"get": Object{
					"parameters": Array{
						Object{
							"$ref": "#/components/parameters/someParameter",
						},
					},
				},
			},
		},
		"components": Object{
			"parameters": Object{
				"someParameter": Object{
					"description": "Some parameter",
				},
			},
			"requestBodies": Object{
				"someRequestBody": Object{
					"description": "Some request body",
				},
			},
			"responses": Object{
				"someResponse": Object{
					"description": "Some response",
				},
			},
			"schemas": Object{
				"someSchema": Object{
					"description": "Some schema",
				},
			},
		},
	})
}
