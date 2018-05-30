package openapi2conv_test

import (
	"encoding/json"
	"fmt"
	"github.com/ronniedada/kin-openapi/openapi2"
	"github.com/ronniedada/kin-openapi/openapi2conv"
	"github.com/ronniedada/kin-openapi/openapi3"
	"github.com/ronniedada/kin-test/jsontest"
	"testing"
)

type Object map[string]interface{}

type Array []interface{}

var Examples = []Example{
	{
		V2: Object{
			"info": Object{},
			"paths": Object{
				"/example": Object{
					"delete": Object{
						"description": "example delete",
					},
					"get": Object{
						"description": "example get",
						"parameters": Array{
							Object{
								"in":   "query",
								"name": "x",
							},
							Object{
								"in":     "body",
								"name":   "body",
								"schema": Object{},
							},
						},
						"responses": Object{
							"default": Object{
								"description": "default response",
							},
							"404": Object{
								"description": "404 response",
							},
						},
						"security": Array{
							Object{
								"get_security_0": Array{"scope0", "scope1"},
								"get_security_1": Array{},
							},
						},
					},
					"head": Object{
						"description": "example head",
					},
					"patch": Object{
						"description": "example patch",
					},
					"post": Object{
						"description": "example post",
					},
					"put": Object{
						"description": "example put",
					},
					"options": Object{
						"description": "example options",
					},
				},
			},
			"security": Array{
				Object{
					"default_security_0": Array{"scope0", "scope1"},
					"default_security_1": Array{},
				},
			},
		},
		V3: Object{
			"openapi":    "3.0",
			"info":       Object{},
			"components": Object{},
			"paths": Object{
				"/example": Object{
					"delete": Object{
						"description": "example delete",
					},
					"get": Object{
						"description": "example get",
						"parameters": Array{
							Object{
								"in":   "query",
								"name": "x",
							},
						},
						"body": Object{
							"content": Object{
								"application/json": Object{
									"schema": Object{},
								},
							},
						},
						"responses": Object{
							"default": Object{
								"description": "default response",
							},
							"404": Object{
								"description": "404 response",
							},
						},
						"security": Array{
							Object{
								"get_security_0": Array{"scope0", "scope1"},
								"get_security_1": Array{},
							},
						},
					},
					"head": Object{
						"description": "example head",
					},
					"options": Object{
						"description": "example options",
					},
					"patch": Object{
						"description": "example patch",
					},
					"post": Object{
						"description": "example post",
					},
					"put": Object{
						"description": "example put",
					},
				},
			},
			"security": Array{
				Object{
					"default_security_0": Array{"scope0", "scope1"},
					"default_security_1": Array{},
				},
			},
		},
	},
}

type Example struct {
	V2 interface{}
	V3 interface{}
}

func copyJSON(dest, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("Failed to marshal %T: %v", src, err)
	}
	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal %T: %v", dest, err)
	}
	return nil
}

func Test_openapi2(t *testing.T) {
	for _, example := range Examples {
		swagger2 := &openapi2.Swagger{}
		swagger3 := &openapi3.Swagger{}
		err := copyJSON(swagger2, example.V2)
		if err != nil {
			panic(err)
		}
		err = copyJSON(swagger3, example.V3)
		if err != nil {
			panic(err)
		}
		t.Log("Converting V3 -> V2")
		actualV2, err := openapi2conv.FromV3Swagger(swagger3)
		jsontest.ExpectWithErr(t, actualV2, err).Value(example.V2)

		t.Log("Converting V2 -> V3")
		actualV3, err := openapi2conv.ToV3Swagger(swagger2)
		jsontest.ExpectWithErr(t, actualV3, err).Value(example.V3)
	}
}
