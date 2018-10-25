package openapi3_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"

	"net/url"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestLoadYAML(t *testing.T) {
	spec := []byte(`
openapi: 3.0.0
info:
  title: An API
  version: v1

components:
  schemas:
    NewItem:
      required: [name]
      properties:
        name: {type: string}
        tag: {type: string}
    ErrorModel:
      type: object
      required: [code, message]
      properties:
        code: {type: integer}
        message: {type: string}

paths:
  /items:
    put:
      description: ''
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/NewItem'
      responses:
        default: &defaultResponse # a YAML ref
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
`)

	loader := openapi3.NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromYAMLData(spec)
	require.NoError(t, err)
	require.Equal(t, "An API", doc.Info.Title)
	require.Equal(t, 2, len(doc.Components.Schemas))
	require.Equal(t, 1, len(doc.Paths))
	def := doc.Paths["/items"].Put.Responses.Default().Value
	require.Equal(t, "unexpected error", def.Description)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}

func ExampleSwaggerLoader() {
	source := `{"info":{"description":"An API"}}`
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(source))
	if err != nil {
		panic(err)
	}
	fmt.Print(swagger.Info.Description)
	// Output:
	// An API
}

func TestResolveSchemaRef(t *testing.T) {
	source := []byte(`{"info":{"description":"An API"},"components":{"schemas":{"B":{"type":"string"},"A":{"allOf":[{"$ref":"#/components/schemas/B"}]}}}}`)
	loader := openapi3.NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(source)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)

	require.NoError(t, err)
	refAVisited := doc.Components.Schemas["A"].Value.AllOf[0]
	require.Equal(t, "#/components/schemas/B", refAVisited.Ref)
	require.NotNil(t, refAVisited.Value)
}

func TestResolveSchemaRefWithNullSchemaRef(t *testing.T) {
	source := []byte(`{"info":{"description":"An API"},"paths":{"/foo":{"post":{"requestBody":{"content":{"application/json":{"schema":null}}}}}}}`)
	loader := openapi3.NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(source)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.EqualError(t, err, "Found unresolved ref: ''")
}

func TestResolveResponseExampleRef(t *testing.T) {
	source := []byte(`
openapi: 3.0.1
info:
  title: My API
  version: 1.0.0
components:
  examples:
    test:
      value:
        error: false
paths:
  /:
    get:
      responses:
        200:
          description: A test response
          content:
            application/json:
              examples:
                test:
                  $ref: '#/components/examples/test'`)
	loader := openapi3.NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromYAMLData(source)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	example := doc.Paths["/"].Get.Responses.Get(200).Value.Content.Get("application/json").Examples["test"]
	require.NotNil(t, example.Value)
	require.Equal(t, example.Value.Value.(map[string]interface{})["error"].(bool), false)
}

type sourceExample struct {
	Location *url.URL
	Spec     []byte
}

type multipleSourceSwaggerLoaderExample struct {
	Sources []*sourceExample
}

func (l *multipleSourceSwaggerLoaderExample) LoadSwaggerFromURI(
	loader *openapi3.SwaggerLoader,
	location *url.URL,
) (*openapi3.Swagger, error) {
	source := l.resolveSourceFromURI(location)
	if source == nil {
		return nil, fmt.Errorf("Unsupported URI: '%s'", location.String())
	}
	return loader.LoadSwaggerFromData(source.Spec)
}

func (l *multipleSourceSwaggerLoaderExample) resolveSourceFromURI(location fmt.Stringer) *sourceExample {
	locationString := location.String()
	for _, v := range l.Sources {
		if v.Location.String() == locationString {
			return v
		}
	}
	return nil
}

func TestResolveSchemaExternalRef(t *testing.T) {
	rootLocation := &url.URL{Scheme: "http", Host: "example.com", Path: "spec.json"}
	externalLocation := &url.URL{Scheme: "http", Host: "example.com", Path: "external.json"}
	rootSpec := []byte(fmt.Sprintf(
		`{"info":{"description":"An API"},"components":{"schemas":{"Root":{"allOf":[{"$ref":"%s#/components/schemas/External"}]}}}}`,
		externalLocation.String(),
	))
	externalSpec := []byte(`{"info":{"description":"External Spec"},"components":{"schemas":{"External":{"type":"string"}}}}`)
	multipleSourceLoader := &multipleSourceSwaggerLoaderExample{
		Sources: []*sourceExample{
			{
				Location: rootLocation,
				Spec:     rootSpec,
			},
			{
				Location: externalLocation,
				Spec:     externalSpec,
			},
		},
	}
	loader := &openapi3.SwaggerLoader{
		IsExternalRefsAllowed:  true,
		LoadSwaggerFromURIFunc: multipleSourceLoader.LoadSwaggerFromURI,
	}
	doc, err := loader.LoadSwaggerFromURI(rootLocation)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)

	require.NoError(t, err)
	refRootVisited := doc.Components.Schemas["Root"].Value.AllOf[0]
	require.Equal(t, fmt.Sprintf("%s#/components/schemas/External", externalLocation.String()), refRootVisited.Ref)
	require.NotNil(t, refRootVisited.Value)
}

func TestLoadErrorOnRefMisuse(t *testing.T) {
	spec := []byte(`
openapi: '3.0.0'
servers: [{url: /}]
info:
  title: ''
  version: '1'
components:
  schemas:
    Thing: {type: string}
paths:
  /items:
    put:
      description: ''
      requestBody:
        $ref: '#/components/schemas/Thing'
      responses:
        '201':
          description: ''
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
`)

	loader := openapi3.NewSwaggerLoader()
	_, err := loader.LoadSwaggerFromYAMLData(spec)
	require.Error(t, err)
}

func TestLoadPathParamRef(t *testing.T) {
	spec := []byte(`
openapi: '3.0.0'
info:
  title: ''
  version: '1'
components:
  parameters:
    testParam:
      name: test
      in: query
      schema:
        type: string
paths:
  '/':
    parameters:
      - $ref: '#/components/parameters/testParam'
    get:
      responses:
        '200':
          description: Test call.
`)

	loader := openapi3.NewSwaggerLoader()
	swagger, err := loader.LoadSwaggerFromYAMLData(spec)
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/"].Parameters[0].Value)
}

func TestLoadRequestExampleRef(t *testing.T) {
	spec := []byte(`
openapi: '3.0.0'
info:
  title: ''
  version: '1'
components:
  examples:
    test:
      value:
        hello: world
paths:
  '/':
    post:
      requestBody:
        content:
          application/json:
            examples:
              test:
                $ref: "#/components/examples/test"
      responses:
        '200':
          description: Test call.
`)

	loader := openapi3.NewSwaggerLoader()
	swagger, err := loader.LoadSwaggerFromYAMLData(spec)
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/"].Post.RequestBody.Value.Content.Get("application/json").Examples["test"])
}

func createTestServer(address string, handler http.Handler) *httptest.Server {
	ts := httptest.NewUnstartedServer(handler)
	l, _ := net.Listen("tcp", address)
	ts.Listener.Close()
	ts.Listener = l
	return ts
}
func TestLoadFromRemoteURL(t *testing.T) {

	fs := http.FileServer(http.Dir("testdata"))
	ts := createTestServer("localhost:3000", fs)
	ts.Start()
	defer ts.Close()

	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	url, err := url.Parse("http://localhost:3000/test.openapi.json")
	require.NoError(t, err)

	swagger, err := loader.LoadSwaggerFromURI(url)
	require.NoError(t, err)

	require.Equal(t, "string", swagger.Components.Schemas["TestSchema"].Value.Type)
}

func TestLoadFileWithExternalSchemaRef(t *testing.T) {
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromFile("testfiles/test.openapi.json")
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Schemas["TestSchema"].Value.Type)
}

func TestLoadFromDataWithExternalSchemaRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {},
    "components": {
        "schemas": {
            "TestSchema": {
                "$ref": "components.openapi.json#/components/schemas/CustomTestSchema"
            }
        }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Schemas["TestSchema"].Value.Type)
}

func TestLoadFromDataWithExternalResponseRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {},
    "components": {
        "responses": {
            "TestResponse": {
                "$ref": "components.openapi.json#/components/responses/CustomTestResponse"
            }
        }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Responses["TestResponse"].Value.Description)
}

func TestLoadRequestResponseHeaderRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
      "/test": {
        "post": {
          "responses": {
            "default": {
              "description": "test",
              "headers": {
                "X-TEST-HEADER": {
                  "$ref": "#/components/headers/TestHeader"
                }
              }
            }
          }
        }
      }
    },
    "components": {
      "headers": {
        "TestHeader": {
          "description": "testheader"
        }
      }
    }
}`)

	loader := openapi3.NewSwaggerLoader()
	swagger, err := loader.LoadSwaggerFromData(spec)
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
	require.Equal(t, "testheader", swagger.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
}

func TestLoadFromDataWithExternalParameterRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {},
    "components": {
        "parameters": {
            "TestParameter": {
                "$ref": "components.openapi.json#/components/parameters/CustomTestParameter"
            }
        }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Parameters["TestParameter"].Value.Name)
}

func TestLoadFromDataWithExternalExampleRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {},
    "components": {
        "examples": {
            "TestExample": {
                "$ref": "components.openapi.json#/components/examples/CustomTestExample"
            }
        }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Examples["TestExample"].Value.Description)
}

func TestLoadFromDataWithExternalRequestBodyRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {},
    "components": {
        "requestBodies": {
            "TestRequestBody": {
                "$ref": "components.openapi.json#/components/requestBodies/CustomTestRequestBody"
            }
        }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.RequestBodies["TestRequestBody"].Value.Content)
}

func TestLoadFromDataWithExternalSecuritySchemeRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {},
    "components": {
        "securitySchemes": {
            "TestSecurityScheme": {
                "$ref": "components.openapi.json#/components/securitySchemes/CustomTestSecurityScheme"
            }
        }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.SecuritySchemes["TestSecurityScheme"].Value.Description)
}

func TestLoadFromDataWithExternalHeaderRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {},
    "components": {
        "headers": {
            "TestHeader": {
                "$ref": "components.openapi.json#/components/headers/CustomTestHeader"
            }
        }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Headers["TestHeader"].Value.Description)
}

func TestLoadFromDataWithPathParameterRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
      "/test/{id}": {
        "parameters": [
          {
            "$ref": "components.openapi.json#/components/parameters/CustomTestParameter"
          }
        ]
      }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/test/{id}"].Parameters[0].Value)
}

func TestLoadFromDataWithPathOperationParameterRef(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
      "/test/{id}": {
        "get": {
          "responses": {},
          "parameters": [
            {
              "$ref": "components.openapi.json#/components/parameters/CustomTestParameter"
            }
          ]
        }
      }
    }
}`)
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testfiles/test.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/test/{id}"].Get.Parameters[0].Value)
}
