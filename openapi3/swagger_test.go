package openapi3

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
)

func TestRefsJSON(t *testing.T) {
	loader := NewSwaggerLoader()

	t.Log("Marshal *Swagger to JSON")
	data, err := json.Marshal(spec())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *Swagger from JSON")
	docA := &Swagger{}
	err = json.Unmarshal(specJSON, &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Resolve refs in unmarshalled *Swagger")
	err = loader.ResolveRefsIn(docA, nil)
	require.NoError(t, err)
	t.Log("Resolve refs in marshalled *Swagger")
	docB, err := loader.LoadSwaggerFromData(data)
	require.NoError(t, err)
	require.NotEmpty(t, docB)

	t.Log("Validate *Swagger")
	err = docA.Validate(loader.Context)
	require.NoError(t, err)
	err = docB.Validate(loader.Context)
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := json.Marshal(docA)
	require.NoError(t, err)
	dataB, err := json.Marshal(docB)
	require.NoError(t, err)
	require.JSONEq(t, string(data), string(specJSON))
	require.JSONEq(t, string(data), string(dataA))
	require.JSONEq(t, string(data), string(dataB))
}

func TestRefsYAML(t *testing.T) {
	loader := NewSwaggerLoader()

	t.Log("Marshal *Swagger to YAML")
	data, err := yaml.Marshal(spec())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *Swagger from YAML")
	docA := &Swagger{}
	err = yaml.Unmarshal(specYAML, &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Resolve refs in unmarshalled *Swagger")
	err = loader.ResolveRefsIn(docA, nil)
	require.NoError(t, err)
	t.Log("Resolve refs in marshalled *Swagger")
	docB, err := loader.LoadSwaggerFromData(data)
	require.NoError(t, err)
	require.NotEmpty(t, docB)

	t.Log("Validate *Swagger")
	err = docA.Validate(loader.Context)
	require.NoError(t, err)
	err = docB.Validate(loader.Context)
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := yaml.Marshal(docA)
	require.NoError(t, err)
	dataB, err := yaml.Marshal(docB)
	require.NoError(t, err)
	eqYAML(t, data, specYAML)
	eqYAML(t, data, dataA)
	eqYAML(t, data, dataB)
}

func eqYAML(t *testing.T, expected, actual []byte) {
	var e, a interface{}
	err := yaml.Unmarshal(expected, &e)
	require.NoError(t, err)
	err = yaml.Unmarshal(actual, &a)
	require.NoError(t, err)
	require.Equal(t, e, a)
}

var specYAML = []byte(`
openapi: '3.0'
info:
  title: MyAPI
  version: '0.1'
paths:
  "/hello":
    parameters:
    - "$ref": "#/components/parameters/someParameter"
    post:
      parameters:
      - "$ref": "#/components/parameters/someParameter"
      requestBody:
        "$ref": "#/components/requestBodies/someRequestBody"
      responses:
        '200':
          "$ref": "#/components/responses/someResponse"
components:
  parameters:
    someParameter:
      description: Some parameter
      name: example
      in: query
      schema:
        "$ref": "#/components/schemas/someSchema"
  requestBodies:
    someRequestBody:
      description: Some request body
  responses:
    someResponse:
      description: Some response
  schemas:
    someSchema:
      description: Some schema
  headers:
    otherHeader: {}
    someHeader:
      "$ref": "#/components/headers/otherHeader"
  examples:
    otherExample:
      value:
        name: Some example
    someExample:
      "$ref": "#/components/examples/otherExample"
  securitySchemes:
    otherSecurityScheme:
      description: Some security scheme
      type: apiKey
      in: query
      name: token
    someSecurityScheme:
      "$ref": "#/components/securitySchemes/otherSecurityScheme"
`)

var specJSON = []byte(`
{
  "openapi": "3.0",
  "info": {
    "title": "MyAPI",
    "version": "0.1"
  },
  "paths": {
    "/hello": {
      "parameters": [
        {
          "$ref": "#/components/parameters/someParameter"
        }
      ],
      "post": {
        "parameters": [
          {
            "$ref": "#/components/parameters/someParameter"
          }
        ],
        "requestBody": {
          "$ref": "#/components/requestBodies/someRequestBody"
        },
        "responses": {
          "200": {
            "$ref": "#/components/responses/someResponse"
          }
        }
      }
    }
  },
  "components": {
    "parameters": {
      "someParameter": {
        "description": "Some parameter",
        "name": "example",
        "in": "query",
        "schema": {
          "$ref": "#/components/schemas/someSchema"
        }
      }
    },
    "requestBodies": {
      "someRequestBody": {
        "description": "Some request body"
      }
    },
    "responses": {
      "someResponse": {
        "description": "Some response"
      }
    },
    "schemas": {
      "someSchema": {
        "description": "Some schema"
      }
    },
    "headers": {
      "otherHeader": {},
      "someHeader": {
        "$ref": "#/components/headers/otherHeader"
      }
    },
    "examples": {
      "otherExample": {
        "value": {
          "name": "Some example"
        }
      },
      "someExample": {
        "$ref": "#/components/examples/otherExample"
      }
    },
    "securitySchemes": {
      "otherSecurityScheme": {
        "description": "Some security scheme",
        "type": "apiKey",
        "in": "query",
        "name": "token"
      },
      "someSecurityScheme": {
        "$ref": "#/components/securitySchemes/otherSecurityScheme"
      }
    }
  }
}
`)

func spec() *Swagger {
	parameter := &Parameter{
		Description: "Some parameter",
		Name:        "example",
		In:          "query",
		Schema: &SchemaRef{
			Ref: "#/components/schemas/someSchema",
		},
	}
	requestBody := &RequestBody{
		Description: "Some request body",
	}
	responseDescription := "Some response"
	response := &Response{
		Description: &responseDescription,
	}
	schema := &Schema{
		Description: "Some schema",
	}
	example := map[string]string{"name": "Some example"}
	return &Swagger{
		OpenAPI: "3.0",
		Info: &Info{
			Title:   "MyAPI",
			Version: "0.1",
		},
		Paths: Paths{
			"/hello": &PathItem{
				Post: &Operation{
					Parameters: Parameters{
						{
							Ref:   "#/components/parameters/someParameter",
							Value: parameter,
						},
					},
					RequestBody: &RequestBodyRef{
						Ref:   "#/components/requestBodies/someRequestBody",
						Value: requestBody,
					},
					Responses: Responses{
						"200": &ResponseRef{
							Ref:   "#/components/responses/someResponse",
							Value: response,
						},
					},
				},
				Parameters: Parameters{
					{
						Ref:   "#/components/parameters/someParameter",
						Value: parameter,
					},
				},
			},
		},
		Components: Components{
			Parameters: map[string]*ParameterRef{
				"someParameter": {
					Value: parameter,
				},
			},
			RequestBodies: map[string]*RequestBodyRef{
				"someRequestBody": {
					Value: requestBody,
				},
			},
			Responses: map[string]*ResponseRef{
				"someResponse": {
					Value: response,
				},
			},
			Schemas: map[string]*SchemaRef{
				"someSchema": {
					Value: schema,
				},
			},
			Headers: map[string]*HeaderRef{
				"someHeader": {
					Ref: "#/components/headers/otherHeader",
				},
				"otherHeader": {
					Value: &Header{},
				},
			},
			Examples: map[string]*ExampleRef{
				"someExample": {
					Ref: "#/components/examples/otherExample",
				},
				"otherExample": {
					Value: NewExample(example),
				},
			},
			SecuritySchemes: map[string]*SecuritySchemeRef{
				"someSecurityScheme": {
					Ref: "#/components/securitySchemes/otherSecurityScheme",
				},
				"otherSecurityScheme": {
					Value: &SecurityScheme{
						Description: "Some security scheme",
						Type:        "apiKey",
						In:          "query",
						Name:        "token",
					},
				},
			},
		},
	}
}

func TestValidation(t *testing.T) {
	info := `
info:
  title: "Hello World REST APIs"
  version: "1.0"
`
	paths := `
paths:
  "/api/v2/greetings.json":
    get:
      operationId: listGreetings
      responses:
        200:
          description: "List different greetings"
  "/api/v2/greetings/{id}.json":
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
          example: "greeting"
    get:
      operationId: showGreeting
      responses:
        200:
          description: "Get a single greeting object"
`
	spec := `
openapi: 3.0.2
` + info + paths + `
components:
  schemas:
    GreetingObject:
      properties:
        id:
          type: string
        type:
          type: string
          default: "greeting"
        attributes:
          properties:
            description:
              type: string
`

	tests := map[string]string{
		spec: "",
		strings.Replace(spec, `openapi: 3.0.2`, ``, 1):            "value of openapi must be a non-empty JSON string",
		strings.Replace(spec, `openapi: 3.0.2`, `openapi: ''`, 1): "value of openapi must be a non-empty JSON string",
		strings.Replace(spec, info, ``, 1):                        "invalid info: must be a JSON object",
		strings.Replace(spec, paths, ``, 1):                       "invalid paths: must be a JSON object",
	}

	for spec, expectedErr := range tests {
		t.Run(expectedErr, func(t *testing.T) {
			doc := &Swagger{}
			err := yaml.Unmarshal([]byte(spec), &doc)
			require.NoError(t, err)

			err = doc.Validate(context.Background())
			if expectedErr != "" {
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
