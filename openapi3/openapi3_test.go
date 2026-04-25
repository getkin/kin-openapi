package openapi3_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/oasdiff/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestRefsJSON(t *testing.T) {
	loader := openapi3.NewLoader()

	t.Log("Marshal *T to JSON")
	data, err := json.Marshal(spec())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *T from JSON")
	docA := &openapi3.T{}
	err = json.Unmarshal(specJSON, &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Resolve refs in unmarshaled *T")
	err = loader.ResolveRefsIn(docA, nil)
	require.NoError(t, err)
	t.Log("Resolve refs in marshaled *T")
	docB, err := loader.LoadFromData(data)
	require.NoError(t, err)
	require.NotEmpty(t, docB)

	t.Log("Validate *T")
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
	loader := openapi3.NewLoader()

	t.Log("Marshal *T to YAML")
	data, err := yaml.Marshal(spec())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *T from YAML")
	docA := &openapi3.T{}
	err = yaml.Unmarshal(specYAML, &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Resolve refs in unmarshaled *T")
	err = loader.ResolveRefsIn(docA, nil)
	require.NoError(t, err)
	t.Log("Resolve refs in marshaled *T")
	docB, err := loader.LoadFromData(data)
	require.NoError(t, err)
	require.NotEmpty(t, docB)

	t.Log("Validate *T")
	err = docA.Validate(loader.Context)
	require.NoError(t, err)
	err = docB.Validate(loader.Context)
	require.NoError(t, err)

	t.Log("Ensure representations match")
	dataA, err := yaml.Marshal(docA)
	require.NoError(t, err)
	dataB, err := yaml.Marshal(docB)
	require.NoError(t, err)
	require.YAMLEq(t, string(data), string(specYAML))
	require.YAMLEq(t, string(data), string(dataA))
	require.YAMLEq(t, string(data), string(dataB))
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
      content: {}
  responses:
    someResponse:
      description: Some response
  schemas:
    someSchema:
      description: Some schema
  headers:
    otherHeader:
      schema: {type: string}
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
        "description": "Some request body",
        "content": {}
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
      "otherHeader": {
        "schema": {
          "type": "string"
      	}
      },
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

func spec() *openapi3.T {
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
		Content:     openapi3.NewContent(),
	}
	responseDescription := "Some response"
	response := &openapi3.Response{
		Description: &responseDescription,
	}
	schema := &openapi3.Schema{
		Description: "Some schema",
	}
	example := map[string]string{"name": "Some example"}
	return &openapi3.T{
		OpenAPI: "3.0",
		Info: &openapi3.Info{
			Title:   "MyAPI",
			Version: "0.1",
		},
		Paths: openapi3.NewPaths(
			openapi3.WithPath("/hello", &openapi3.PathItem{
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
					Responses: openapi3.NewResponses(
						openapi3.WithStatus(200, &openapi3.ResponseRef{
							Ref:   "#/components/responses/someResponse",
							Value: response,
						}),
					),
				},
				Parameters: openapi3.Parameters{
					{
						Ref:   "#/components/parameters/someParameter",
						Value: parameter,
					},
				},
			}),
		),
		Components: &openapi3.Components{
			Parameters: openapi3.ParametersMap{
				"someParameter": {Value: parameter},
			},
			RequestBodies: openapi3.RequestBodies{
				"someRequestBody": {Value: requestBody},
			},
			Responses: openapi3.ResponseBodies{
				"someResponse": {Value: response},
			},
			Schemas: openapi3.Schemas{
				"someSchema": {Value: schema},
			},
			Headers: openapi3.Headers{
				"someHeader":  {Ref: "#/components/headers/otherHeader"},
				"otherHeader": {Value: &openapi3.Header{openapi3.Parameter{Schema: &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}}}},
			},
			Examples: openapi3.Examples{
				"someExample":  {Ref: "#/components/examples/otherExample"},
				"otherExample": {Value: openapi3.NewExample(example)},
			},
			SecuritySchemes: openapi3.SecuritySchemes{
				"someSecurityScheme": {Ref: "#/components/securitySchemes/otherSecurityScheme"},
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

func TestValidation(t *testing.T) {
	version := `
openapi: 3.0.2
`
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
	externalDocs := `
externalDocs:
  url: https://root-ext-docs.com
`
	tags := `
tags:
  - name: "pet"
    externalDocs:
      url: https://tags-ext-docs.com
`
	spec := version + info + paths + externalDocs + tags + `
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

	tests := []struct {
		name        string
		spec        string
		expectedErr string
	}{
		{
			name: "no errors",
			spec: spec,
		},
		{
			name:        "version is missing",
			spec:        strings.Replace(spec, version, "", 1),
			expectedErr: "value of openapi must be a non-empty string",
		},
		{
			name:        "version is empty string",
			spec:        strings.Replace(spec, version, "openapi: ''", 1),
			expectedErr: "value of openapi must be a non-empty string",
		},
		{
			name:        "info section is missing",
			spec:        strings.Replace(spec, info, ``, 1),
			expectedErr: "invalid info: must be an object",
		},
		{
			name:        "paths section is missing",
			spec:        strings.Replace(spec, paths, ``, 1),
			expectedErr: "invalid paths: must be an object",
		},
		{
			name: "externalDocs section is invalid",
			spec: strings.Replace(spec, externalDocs,
				strings.ReplaceAll(externalDocs, "url: https://root-ext-docs.com", "url: ''"), 1),
			expectedErr: "invalid external docs: url is required",
		},
		{
			name: "tags section is invalid",
			spec: strings.Replace(spec, tags,
				strings.ReplaceAll(tags, "url: https://tags-ext-docs.com", "url: ''"), 1),
			expectedErr: "invalid tags: invalid external docs: url is required",
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			doc := &openapi3.T{}
			err := yaml.Unmarshal([]byte(tt.spec), &doc)
			require.NoError(t, err)

			err = doc.Validate(t.Context())
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAddRemoveServer(t *testing.T) {
	testServerLines := []*openapi3.Server{{URL: "test0.com"}, {URL: "test1.com"}, {URL: "test3.com"}}

	doc3 := &openapi3.T{
		OpenAPI:    "3.0.3",
		Components: &openapi3.Components{},
	}

	assert.Empty(t, doc3.Servers)

	doc3.AddServer(&openapi3.Server{URL: "testserver1.com"})

	assert.NotEmpty(t, doc3.Servers)
	assert.Len(t, doc3.Servers, 1)

	doc3.Servers = openapi3.Servers{}

	assert.Empty(t, doc3.Servers)

	doc3.AddServers(testServerLines[0], testServerLines[1], testServerLines[2])

	assert.NotEmpty(t, doc3.Servers)
	assert.Len(t, doc3.Servers, 3)

	doc3.Servers = openapi3.Servers{}

	doc3.AddServers(testServerLines...)

	assert.NotEmpty(t, doc3.Servers)
	assert.Len(t, doc3.Servers, 3)

	doc3.Servers = openapi3.Servers{}
}

func TestOpenAPIMajorMinor(t *testing.T) {
	var doc *openapi3.T
	require.Equal(t, "", doc.OpenAPIMajorMinor())
	require.False(t, doc.IsOpenAPI30())
	require.False(t, doc.IsOpenAPI31OrLater())

	doc = &openapi3.T{}
	require.Equal(t, "", doc.OpenAPIMajorMinor())
	require.False(t, doc.IsOpenAPI30())
	require.False(t, doc.IsOpenAPI31OrLater())

	semvers := []string{"3", "3.0", "3.0.0", "3.0.1", "3.0.2", "3.0.3", "3.0.4", "3.1", "3.1.0", "3.1.1", "3.1.2", "3.2", "3.2.0"}
	mms := []string{"3.0", "3.0", "3.0", "3.0", "3.0", "3.0", "3.0", "3.1", "3.1", "3.1", "3.1", "3.2", "3.2"}
	three0s := []bool{true, true, true, true, true, true, true, false, false, false, false, false, false}
	three1plusses := []bool{false, false, false, false, false, false, false, true, true, true, true, true, true}
	for i := range len(semvers) {
		t.Run(fmt.Sprintf("openapi:%s", semvers[i]), func(t *testing.T) {
			t.Parallel()
			doc := &openapi3.T{OpenAPI: semvers[i]}
			require.Equal(t, mms[i], doc.OpenAPIMajorMinor())
			require.Equal(t, three0s[i], doc.IsOpenAPI30())
			require.Equal(t, three1plusses[i], doc.IsOpenAPI31OrLater())
		})
	}
}
