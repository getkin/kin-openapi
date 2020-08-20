package openapi3_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
)

func TestRefsJSON(t *testing.T) {
	loader1 := openapi3.NewSwaggerLoader()
	loader2 := openapi3.NewSwaggerLoader()

	t.Log("Marshal *openapi3.Swagger to JSON")
	data, err := json.Marshal(spec())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *openapi3.Swagger from JSON")
	docA := &openapi3.Swagger{}
	err = json.Unmarshal([]byte(specJSON), &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Resolve refs in unmarshalled *openapi3.Swagger")
	err = loader1.ResolveRefsIn(docA, nil)
	require.NoError(t, err)
	t.Log("Resolve refs in marshalled *openapi3.Swagger")
	docB, err := loader2.LoadSwaggerFromData(data)
	require.NoError(t, err)
	require.NotEmpty(t, docB)

	t.Log("Validate *openapi3.Swagger")
	err = docA.Validate(loader1.Context)
	require.NoError(t, err)
	err = docB.Validate(loader2.Context)
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := json.Marshal(docA)
	require.NoError(t, err)
	dataB, err := json.Marshal(docB)
	require.NoError(t, err)
	require.JSONEq(t, string(data), specJSON)
	require.JSONEq(t, string(data), string(dataA))
	require.JSONEq(t, string(data), string(dataB))
}

func TestRefsYAML(t *testing.T) {
	loader := openapi3.NewSwaggerLoader()

	t.Log("Marshal *openapi3.Swagger to YAML")
	data, err := yaml.Marshal(spec())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *openapi3.Swagger from YAML")
	docA := &openapi3.Swagger{}
	err = yaml.Unmarshal([]byte(specYAML), &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Resolve refs in unmarshalled *openapi3.Swagger")
	err = loader.ResolveRefsIn(docA, nil)
	require.NoError(t, err)
	t.Log("Resolve refs in marshalled *openapi3.Swagger")
	docB, err := loader.LoadSwaggerFromData(data)
	require.NoError(t, err)
	require.NotEmpty(t, docB)

	t.Log("Validate *openapi3.Swagger")
	err = docA.Validate(loader.Context)
	require.NoError(t, err)
	err = docB.Validate(loader.Context)
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := yaml.Marshal(docA)
	require.NoError(t, err)
	dataB, err := yaml.Marshal(docB)
	require.NoError(t, err)
	eqYAML(t, data, []byte(specYAML))
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

var specYAML = `
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
`

var specJSON = `
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
`

func spec() *openapi3.Swagger {
	parameter := &openapi3.Parameter{
		Description: "Some parameter",
		Name:        "example",
		In:          "query",
		Schema: &openapi3.SchemaRef{
			Ref: "#/components/schemas/someSchema",
		},
	}
	requestBody := &openapi3.RequestBody{
		Description: "Some request body",
	}
	responseDescription := "Some response"
	response := &openapi3.Response{
		Description: &responseDescription,
	}
	schema := &openapi3.Schema{
		Description: "Some schema",
	}
	example := map[string]string{"name": "Some example"}
	return &openapi3.Swagger{
		OpenAPI: "3.0",
		Info: &openapi3.Info{
			Title:   "MyAPI",
			Version: "0.1",
		},
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
			Headers: map[string]*openapi3.HeaderRef{
				"someHeader": {
					Ref: "#/components/headers/otherHeader",
				},
				"otherHeader": {
					Value: &openapi3.Header{},
				},
			},
			Examples: map[string]*openapi3.ExampleRef{
				"someExample": {
					Ref: "#/components/examples/otherExample",
				},
				"otherExample": {
					Value: openapi3.NewExample(example),
				},
			},
			SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{
				"someSecurityScheme": {
					Ref: "#/components/securitySchemes/otherSecurityScheme",
				},
				"otherSecurityScheme": {
					Value: &openapi3.SecurityScheme{
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

// TestValidation tests validation of properties in the root of the OpenAPI
// file.
func TestValidation(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError error
	}{
		{
			"when no OpenAPI property is supplied",
			`
info:
  title: "Hello World REST APIs"
  version: "1.0"
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
`,
			errors.New("value of openapi must be a non-empty JSON string"),
		},
		{
			"when an empty OpenAPI property is supplied",
			`
openapi: ''
info:
  title: "Hello World REST APIs"
  version: "1.0"
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
`,
			errors.New("value of openapi must be a non-empty JSON string"),
		},
		{
			"when the Info property is not supplied",
			`
openapi: '1.0'
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
`,
			errors.New("invalid info: must be a JSON object"),
		},
		{
			"when the Paths property is not supplied",
			`
openapi: '1.0'
info:
  title: "Hello World REST APIs"
  version: "1.0"
`,
			errors.New("invalid paths: must be a JSON object"),
		},
		{
			"when a valid spec is supplied",
			`
openapi: 3.0.2
info:
  title: "Hello World REST APIs"
  version: "1.0"
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
`,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc := &openapi3.Swagger{}
			err := yaml.Unmarshal([]byte(test.input), &doc)
			require.NoError(t, err)

			c := context.Background()
			validationErr := doc.Validate(c)

			require.Equal(t, test.expectedError, validationErr, "expected errors (or lack of) to match")
		})
	}
}
