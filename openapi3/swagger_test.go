package openapi3_test

import (
	"context"
	"encoding/json"
	"github.com/jban332/kin-openapi/openapi3"
	"github.com/jban332/kin-core/jsontest"
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

	t.Log("Resolve refs in unmarshalled *openapi3.Swagger")
	swagger.ResolveRefs()

	t.Log("Validate unmarshalled *openapi3.Swagger")
	err = swagger.Validate(context.TODO())
	jsontest.ExpectErr(t, err).Err(nil)
}

func TestRefs(t *testing.T) {
	parameter := &openapi3.Parameter{
		Description: "Some parameter",
		Name:        "example",
		In:          "query",
	}
	requestBody := &openapi3.RequestBody{
		Description: "Some request body",
	}
	response := &openapi3.Response{
		Description: "Some response",
	}
	schema := &openapi3.Schema{
		Description: "Some schema",
	}
	swagger := &openapi3.Swagger{
		OpenAPI: "3.0",
		Paths: openapi3.Paths{
			"/hello": &openapi3.PathItem{
				Post: &openapi3.Operation{
					Parameters: openapi3.Parameters{
						{
							Ref:   "#/components/parameters/someParameter",
							Value: parameter,
						},
					},
					RequestBody: &openapi3.RequestBodyRef{
						Ref:   "#/components/requestBodies/someRequestBody",
						Value: requestBody,
					},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Ref:   "#/components/responses/someResponse",
							Value: response,
						},
					},
				},
				Parameters: openapi3.Parameters{
					{
						Ref:   "#/components/parameters/someParameter",
						Value: parameter,
					},
				},
			},
		},
		Components: openapi3.Components{
			Parameters: map[string]*openapi3.ParameterRef{
				"someParameter": {
					Value: parameter,
				},
			},
			RequestBodies: map[string]*openapi3.RequestBodyRef{
				"someRequestBody": {
					Value: requestBody,
				},
			},
			Responses: map[string]*openapi3.ResponseRef{
				"someResponse": {
					Value: response,
				},
			},
			Schemas: map[string]*openapi3.SchemaRef{
				"someSchema": {
					Value: schema,
				},
			},
		},
	}
	expect(t, swagger, Object{
		"openapi": "3.0",
		"info":    Object{},
		"paths": Object{
			"/hello": Object{
				"parameters": Array{
					Object{
						"$ref": "#/components/parameters/someParameter",
					},
				},
				"post": Object{
					"parameters": Array{
						Object{
							"$ref": "#/components/parameters/someParameter",
						},
					},
					"body": Object{
						"$ref": "#/components/requestBodies/someRequestBody",
					},
					"responses": Object{
						"200": Object{
							"$ref": "#/components/responses/someResponse",
						},
					},
				},
			},
		},
		"components": Object{
			"parameters": Object{
				"someParameter": Object{
					"description": "Some parameter",
					"name":        "example",
					"in":          "query",
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
