package openapi3

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

const addr = "localhost:7965"

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

	loader := NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(spec)
	require.NoError(t, err)
	require.Equal(t, "An API", doc.Info.Title)
	require.Equal(t, 2, len(doc.Components.Schemas))
	require.Equal(t, 1, len(doc.Paths))
	def := doc.Paths["/items"].Put.Responses.Default().Value
	desc := "unexpected error"
	require.Equal(t, &desc, def.Description)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}

func ExampleSwaggerLoader() {
	source := `{"info":{"description":"An API"}}`
	swagger, err := NewSwaggerLoader().LoadSwaggerFromData([]byte(source))
	if err != nil {
		panic(err)
	}
	fmt.Print(swagger.Info.Description)
	// Output:
	// An API
}

func TestResolveSchemaRef(t *testing.T) {
	source := []byte(`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1",description":"An API"},"paths":{},"components":{"schemas":{"B":{"type":"string"},"A":{"allOf":[{"$ref":"#/components/schemas/B"}]}}}}`)
	loader := NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(source)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)

	require.NoError(t, err)
	refAVisited := doc.Components.Schemas["A"].Value.AllOf[0]
	require.Equal(t, "#/components/schemas/B", refAVisited.Ref)
	require.NotNil(t, refAVisited.Value)
}

func TestResolveSchemaRefWithNullSchemaRef(t *testing.T) {
	source := []byte(`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"An API"},"paths":{"/foo":{"post":{"requestBody":{"content":{"application/json":{"schema":null}}}}}}}`)
	loader := NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(source)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.EqualError(t, err, `invalid paths: found unresolved ref: ""`)
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
	loader := NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(source)
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
	loader *SwaggerLoader,
	location *url.URL,
) (*Swagger, error) {
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
		`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"An API"},"paths":{},"components":{"schemas":{"Root":{"allOf":[{"$ref":"%s#/components/schemas/External"}]}}}}`,
		externalLocation.String(),
	))
	externalSpec := []byte(`{"openapi":"3.0.0","info":{"title":"MyAPI","version":"0.1","description":"External Spec"},"paths":{},"components":{"schemas":{"External":{"type":"string"}}}}`)
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
	loader := &SwaggerLoader{
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

	loader := NewSwaggerLoader()
	_, err := loader.LoadSwaggerFromData(spec)
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

	loader := NewSwaggerLoader()
	swagger, err := loader.LoadSwaggerFromData(spec)
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

	loader := NewSwaggerLoader()
	swagger, err := loader.LoadSwaggerFromData(spec)
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/"].Post.RequestBody.Value.Content.Get("application/json").Examples["test"])
}

func createTestServer(handler http.Handler) *httptest.Server {
	ts := httptest.NewUnstartedServer(handler)
	l, _ := net.Listen("tcp", addr)
	ts.Listener.Close()
	ts.Listener = l
	return ts
}

func TestLoadFromRemoteURL(t *testing.T) {

	fs := http.FileServer(http.Dir("testdata"))
	ts := createTestServer(fs)
	ts.Start()
	defer ts.Close()

	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	url, err := url.Parse("http://" + addr + "/test.openapi.json")
	require.NoError(t, err)

	swagger, err := loader.LoadSwaggerFromURI(url)
	require.NoError(t, err)

	require.Equal(t, "string", swagger.Components.Schemas["TestSchema"].Value.Type)
}

func TestLoadFileWithExternalSchemaRef(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromFile("testdata/testref.openapi.json")
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Schemas["AnotherTestSchema"].Value.Type)
}

func TestLoadFileWithExternalSchemaRefSingleComponent(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromFile("testdata/testrefsinglecomponent.openapi.json")
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Responses["SomeResponse"])
	desc := "this is a single response definition"
	require.Equal(t, &desc, swagger.Components.Responses["SomeResponse"].Value.Description)
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

	loader := NewSwaggerLoader()
	swagger, err := loader.LoadSwaggerFromData(spec)
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
	require.Equal(t, "testheader", swagger.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
}

func TestLoadFromDataWithExternalRequestResponseHeaderRemoteRef(t *testing.T) {
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
                                "$ref": "http://` + addr + `/components.openapi.json#/components/headers/CustomTestHeader"
                            }
                        }
                    }
                }
            }
        }
    }
}`)

	fs := http.FileServer(http.Dir("testdata"))
	ts := createTestServer(fs)
	ts.Start()
	defer ts.Close()

	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testdata/testfilename.openapi.json"})
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
	require.Equal(t, "description", swagger.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
}

func TestLoadYamlFile(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromFile("testdata/test.openapi.yml")
	require.NoError(t, err)

	require.Equal(t, "OAI Specification in YAML", swagger.Info.Title)
}

func TestLoadYamlFileWithExternalSchemaRef(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromFile("testdata/testref.openapi.yml")
	require.NoError(t, err)

	require.NotNil(t, swagger.Components.Schemas["AnotherTestSchema"].Value.Type)
}

func TestLoadYamlFileWithExternalPathRef(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromFile("testdata/pathref.openapi.yml")
	require.NoError(t, err)

	require.NotNil(t, swagger.Paths["/test"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Type)
	require.Equal(t, "string", swagger.Paths["/test"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Type)
}

func TestResolveResponseLinkRef(t *testing.T) {
	source := []byte(`
openapi: 3.0.1
info:
  title: My API
  version: 1.0.0
components:
  links:
    Father:
        description: link to to the father
        operationId: getUserById
        parameters:
          "id": "$response.body#/fatherId"
paths:
  /users/{id}:
    get:
      operationId: getUserById,
      parameters:
        - name: id,
          in: path
          schema:
            type: string
      responses:
        200:
          description: A test response
          content:
            application/json:
          links:
            father:
              $ref: '#/components/links/Father'
`)
	loader := NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(source)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	response := doc.Paths[`/users/{id}`].Get.Responses.Get(200).Value
	link := response.Links[`father`].Value
	require.NotNil(t, link)
	require.Equal(t, "getUserById", link.OperationID)
	require.Equal(t, "link to to the father", link.Description)
}

func TestResolveNonComponentsRef(t *testing.T) {
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
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
    post:
      description: ''
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/paths/~1items/put/requestBody/content/application~1json/schema'
      responses:
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
`)

	loader := NewSwaggerLoader()
	doc, err := loader.LoadSwaggerFromData(spec)
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
