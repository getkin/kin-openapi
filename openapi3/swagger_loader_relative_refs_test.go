package openapi3

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type refTestDataEntry struct {
	name            string
	contentTemplate string
	testFunc        func(t *testing.T, swagger *Swagger)
}

type refTestDataEntryWithErrorMessage struct {
	name            string
	contentTemplate string
	errorMessage    *string
	testFunc        func(t *testing.T, swagger *Swagger)
}

var refTestDataEntries = []refTestDataEntry{
	{
		name:            "SchemaRef",
		contentTemplate: externalSchemaRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.Schemas["TestSchema"].Value.Type)
			require.Equal(t, "string", swagger.Components.Schemas["TestSchema"].Value.Type)
		},
	},
	{
		name:            "ResponseRef",
		contentTemplate: externalResponseRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			desc := "description"
			require.Equal(t, &desc, swagger.Components.Responses["TestResponse"].Value.Description)
		},
	},
	{
		name:            "ParameterRef",
		contentTemplate: externalParameterRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.Parameters["TestParameter"].Value.Name)
			require.Equal(t, "id", swagger.Components.Parameters["TestParameter"].Value.Name)
		},
	},
	{
		name:            "ExampleRef",
		contentTemplate: externalExampleRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.Examples["TestExample"].Value.Description)
			require.Equal(t, "description", swagger.Components.Examples["TestExample"].Value.Description)
		},
	},
	{
		name:            "RequestBodyRef",
		contentTemplate: externalRequestBodyRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.RequestBodies["TestRequestBody"].Value.Content)
		},
	},
	{
		name:            "SecuritySchemeRef",
		contentTemplate: externalSecuritySchemeRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.SecuritySchemes["TestSecurityScheme"].Value.Description)
			require.Equal(t, "description", swagger.Components.SecuritySchemes["TestSecurityScheme"].Value.Description)
		},
	},
	{
		name:            "ExternalHeaderRef",
		contentTemplate: externalHeaderRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.Headers["TestHeader"].Value.Description)
			require.Equal(t, "description", swagger.Components.Headers["TestHeader"].Value.Description)
		},
	},
	{
		name:            "PathParameterRef",
		contentTemplate: externalPathParameterRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test/{id}"].Parameters[0].Value.Name)
			require.Equal(t, "id", swagger.Paths["/test/{id}"].Parameters[0].Value.Name)
		},
	},
	{
		name:            "PathOperationParameterRef",
		contentTemplate: externalPathOperationParameterRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test/{id}"].Get.Parameters[0].Value)
			require.Equal(t, "id", swagger.Paths["/test/{id}"].Get.Parameters[0].Value.Name)
		},
	},
	{
		name:            "PathOperationRequestBodyRef",
		contentTemplate: externalPathOperationRequestBodyRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test"].Post.RequestBody.Value)
			require.NotNil(t, swagger.Paths["/test"].Post.RequestBody.Value.Content)
		},
	},
	{
		name:            "PathOperationResponseRef",
		contentTemplate: externalPathOperationResponseRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test"].Post.Responses["default"].Value)
			desc := "description"
			require.Equal(t, &desc, swagger.Paths["/test"].Post.Responses["default"].Value.Description)
		},
	},
	{
		name:            "PathOperationParameterSchemaRef",
		contentTemplate: externalPathOperationParameterSchemaRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test/{id}"].Get.Parameters[0].Value.Schema.Value)
			require.Equal(t, "string", swagger.Paths["/test/{id}"].Get.Parameters[0].Value.Schema.Value.Type)
			require.Equal(t, "id", swagger.Paths["/test/{id}"].Get.Parameters[0].Value.Name)
		},
	},

	{
		name:            "PathOperationParameterRefWithContentInQuery",
		contentTemplate: externalPathOperationParameterWithContentInQueryTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			schemaRef := swagger.Paths["/test/{id}"].Get.Parameters[0].Value.Content["application/json"].Schema
			require.NotNil(t, schemaRef.Value)
			require.Equal(t, "string", schemaRef.Value.Type)
		},
	},

	{
		name:            "PathOperationRequestBodyExampleRef",
		contentTemplate: externalPathOperationRequestBodyExampleRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test"].Post.RequestBody.Value.Content["application/json"].Examples["application/json"].Value)
			require.Equal(t, "description", swagger.Paths["/test"].Post.RequestBody.Value.Content["application/json"].Examples["application/json"].Value.Description)
		},
	},
	{
		name:            "PathOperationReqestBodyContentSchemaRef",
		contentTemplate: externalPathOperationReqestBodyContentSchemaRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test"].Post.RequestBody.Value.Content["application/json"].Schema.Value)
			require.Equal(t, "string", swagger.Paths["/test"].Post.RequestBody.Value.Content["application/json"].Schema.Value.Type)
		},
	},
	{
		name:            "PathOperationResponseExampleRef",
		contentTemplate: externalPathOperationResponseExampleRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test"].Post.Responses["default"].Value)
			desc := "testdescription"
			require.Equal(t, &desc, swagger.Paths["/test"].Post.Responses["default"].Value.Description)
			require.Equal(t, "description", swagger.Paths["/test"].Post.Responses["default"].Value.Content["application/json"].Examples["application/json"].Value.Description)
		},
	},
	{
		name:            "PathOperationResponseSchemaRef",
		contentTemplate: externalPathOperationResponseSchemaRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test"].Post.Responses["default"].Value)
			desc := "testdescription"
			require.Equal(t, &desc, swagger.Paths["/test"].Post.Responses["default"].Value.Description)
			require.Equal(t, "string", swagger.Paths["/test"].Post.Responses["default"].Value.Content["application/json"].Schema.Value.Type)
		},
	},
	{
		name:            "ComponentHeaderSchemaRef",
		contentTemplate: externalComponentHeaderSchemaRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.Headers["TestHeader"].Value)
			require.Equal(t, "string", swagger.Components.Headers["TestHeader"].Value.Schema.Value.Type)
		},
	},
	{
		name:            "RequestResponseHeaderRef",
		contentTemplate: externalRequestResponseHeaderRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
			require.Equal(t, "description", swagger.Paths["/test"].Post.Responses["default"].Value.Headers["X-TEST-HEADER"].Value.Description)
		},
	},
}

var refTestDataEntriesResponseError = []refTestDataEntryWithErrorMessage{
	{
		name:            "CannotContainBothSchemaAndContentInAParameter",
		contentTemplate: externalCannotContainBothSchemaAndContentInAParameter,
		errorMessage:    &(&struct{ x string }{"cannot contain both schema and content in a parameter"}).x,
		testFunc: func(t *testing.T, swagger *Swagger) {
		},
	},
}

func TestLoadFromDataWithExternalRef(t *testing.T) {
	for _, td := range refTestDataEntries {
		t.Logf("testcase '%s'", td.name)

		spec := []byte(fmt.Sprintf(td.contentTemplate, "components.openapi.json"))
		loader := NewSwaggerLoader()
		loader.IsExternalRefsAllowed = true
		swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testdata/testfilename.openapi.json"})
		require.NoError(t, err)
		td.testFunc(t, swagger)
	}
}

func TestLoadFromDataWithExternalRefResponseError(t *testing.T) {
	for _, td := range refTestDataEntriesResponseError {
		t.Logf("testcase '%s'", td.name)

		spec := []byte(fmt.Sprintf(td.contentTemplate, "components.openapi.json"))
		loader := NewSwaggerLoader()
		loader.IsExternalRefsAllowed = true
		swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testdata/testfilename.openapi.json"})
		require.EqualError(t, err, *td.errorMessage)
		td.testFunc(t, swagger)
	}
}

func TestLoadFromDataWithExternalNestedRef(t *testing.T) {
	for _, td := range refTestDataEntries {
		t.Logf("testcase '%s'", td.name)

		spec := []byte(fmt.Sprintf(td.contentTemplate, "nesteddir/nestedcomponents.openapi.json"))
		loader := NewSwaggerLoader()
		loader.IsExternalRefsAllowed = true
		swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testdata/testfilename.openapi.json"})
		require.NoError(t, err)
		td.testFunc(t, swagger)
	}
}

const externalSchemaRefTemplate = `
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
                "$ref": "%s#/components/schemas/CustomTestSchema"
            }
        }
    }
}`

const externalResponseRefTemplate = `
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
                "$ref": "%s#/components/responses/CustomTestResponse"
            }
        }
    }
}`

const externalParameterRefTemplate = `
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
                "$ref": "%s#/components/parameters/CustomTestParameter"
            }
        }
    }
}`

const externalExampleRefTemplate = `
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
                "$ref": "%s#/components/examples/CustomTestExample"
            }
        }
    }
}`

const externalRequestBodyRefTemplate = `
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
                "$ref": "%s#/components/requestBodies/CustomTestRequestBody"
            }
        }
    }
}`

const externalSecuritySchemeRefTemplate = `
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
                "$ref": "%s#/components/securitySchemes/CustomTestSecurityScheme"
            }
        }
    }
}`

const externalHeaderRefTemplate = `
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
                "$ref": "%s#/components/headers/CustomTestHeader"
            }
        }
    }
}`

const externalPathParameterRefTemplate = `
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
                    "$ref": "%s#/components/parameters/CustomTestParameter"
                }
            ]
        }
    }
}`

const externalPathOperationParameterRefTemplate = `
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
                        "$ref": "%s#/components/parameters/CustomTestParameter"
                    }
                ]
            }
        }
    }
}`

const externalPathOperationRequestBodyRefTemplate = `
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
        "/test": {
            "post": {
                "responses": {},
                "requestBody": {
                    "$ref": "%s#/components/requestBodies/CustomTestRequestBody"
                }
            }
        }
    }
}`

const externalPathOperationResponseRefTemplate = `
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
                        "$ref": "%s#/components/responses/CustomTestResponse"
                    }
                }
            }
        }
    }
}`

const externalPathOperationParameterSchemaRefTemplate = `
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
                        "$ref": "#/components/parameters/CustomTestParameter"
                    }
                ]
            }
        }
    },
    "components": {
        "parameters": {
            "CustomTestParameter": {
                "name": "id",
                "in": "header",
                "schema": {
                    "$ref": "%s#/components/schemas/CustomTestSchema"
                }
            }
        }
    }
}`

const externalPathOperationParameterWithContentInQueryTemplate = `
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
                        "$ref": "#/components/parameters/CustomTestParameter"
                    }
                ]
            }
        }
    },
    "components": {
        "parameters": {
            "CustomTestParameter": {
                "content": {
                    "application/json": {
                        "schema": {
                            "$ref": "%s#/components/schemas/CustomTestSchema"
                        }
                    }
                }
            }
        }
    }
}`

const externalCannotContainBothSchemaAndContentInAParameter = `
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
                        "$ref": "#/components/parameters/CustomTestParameter"
                    }
                ]
            }
        }
    },
    "components": {
        "parameters": {
            "CustomTestParameter": {
                "content": {
                    "application/json": {
                        "schema": {
                            "$ref": "%s#/components/schemas/CustomTestSchema"
                        }
                    }
                },
                "schema": {
                    "$ref": "%s#/components/schemas/CustomTestSchema"
                }
            }
        }
    }
}`

const externalPathOperationRequestBodyExampleRefTemplate = `
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
        "/test": {
            "post": {
                "responses": {},
                "requestBody": {
                    "$ref": "#/components/requestBodies/CustomTestRequestBody"
                }
            }
        }
    },
    "components": {
        "requestBodies": {
            "CustomTestRequestBody": {
                "content": {
                    "application/json": {
                        "examples": {
                            "application/json": {
                                "$ref": "%s#/components/examples/CustomTestExample"
                            }
                        }
                    }
                }
            }
        }
    }
}`

const externalPathOperationReqestBodyContentSchemaRefTemplate = `
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
        "/test": {
            "post": {
                "responses": {},
                "requestBody": {
                    "$ref": "#/components/requestBodies/CustomTestRequestBody"
                }
            }
        }
    },
    "components": {
        "requestBodies": {
            "CustomTestRequestBody": {
                "content": {
                    "application/json": {
                        "schema": {
                            "$ref": "%s#/components/schemas/CustomTestSchema"
                        }
                    }
                }
            }
        }
    }
}`

const externalPathOperationResponseExampleRefTemplate = `
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
                        "$ref": "#/components/responses/CustomTestResponse"
                    }
                }
            }
        }
    },
    "components": {
        "responses": {
            "CustomTestResponse": {
                "description": "testdescription",
                "content": {
                    "application/json": {
                        "examples": {
                            "application/json": {
                                "$ref": "%s#/components/examples/CustomTestExample"
                            }
                        }
                    }
                }
            }
        }
    }
}`

const externalPathOperationResponseSchemaRefTemplate = `
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
                        "$ref": "#/components/responses/CustomTestResponse"
                    }
                }
            }
        }
    },
    "components": {
        "responses": {
            "CustomTestResponse": {
                "description": "testdescription",
                "content": {
                    "application/json": {
                        "schema": {
                            "$ref": "%s#/components/schemas/CustomTestSchema"
                        }
                    }
                }
            }
        }
    }
}`

const externalComponentHeaderSchemaRefTemplate = `
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
                "$ref": "#/components/headers/CustomTestHeader"
            },
            "CustomTestHeader": {
                "name": "X-TEST-HEADER",
                "in": "header",
                "schema": {
                    "$ref": "%s#/components/schemas/CustomTestSchema"
                }
            }
        }
    }
}`

const externalRequestResponseHeaderRefTemplate = `
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
                                "$ref": "%s#/components/headers/CustomTestHeader"
                            }
                        }
                    }
                }
            }
        }
    }
}`

// Relative Schema Documents Tests
var relativeDocRefsTestDataEntries = []refTestDataEntry{
	{
		name:            "SchemaRef",
		contentTemplate: relativeSchemaDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.Schemas["TestSchema"].Value.Type)
			require.Equal(t, "string", swagger.Components.Schemas["TestSchema"].Value.Type)
		},
	},
	{
		name:            "ResponseRef",
		contentTemplate: relativeResponseDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			desc := "description"
			require.Equal(t, &desc, swagger.Components.Responses["TestResponse"].Value.Description)
		},
	},
	{
		name:            "ParameterRef",
		contentTemplate: relativeParameterDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.Parameters["TestParameter"].Value.Name)
			require.Equal(t, "param", swagger.Components.Parameters["TestParameter"].Value.Name)
			require.Equal(t, true, swagger.Components.Parameters["TestParameter"].Value.Required)
		},
	},
	{
		name:            "ExampleRef",
		contentTemplate: relativeExampleDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, "param", swagger.Components.Examples["TestExample"].Value.Summary)
			require.NotNil(t, "param", swagger.Components.Examples["TestExample"].Value.Value)
			require.Equal(t, "An example", swagger.Components.Examples["TestExample"].Value.Summary)
		},
	},
	{
		name:            "RequestRef",
		contentTemplate: relativeRequestDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, "param", swagger.Components.RequestBodies["TestRequestBody"].Value.Description)
			require.Equal(t, "example request", swagger.Components.RequestBodies["TestRequestBody"].Value.Description)
		},
	},
	{
		name:            "HeaderRef",
		contentTemplate: relativeHeaderDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, "param", swagger.Components.Headers["TestHeader"].Value.Description)
			require.Equal(t, "description", swagger.Components.Headers["TestHeader"].Value.Description)
		},
	},
	{
		name:            "HeaderRef",
		contentTemplate: relativeHeaderDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, "param", swagger.Components.Headers["TestHeader"].Value.Description)
			require.Equal(t, "description", swagger.Components.Headers["TestHeader"].Value.Description)
		},
	},
	{
		name:            "SecuritySchemeRef",
		contentTemplate: relativeSecuritySchemeDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Components.SecuritySchemes["TestSecurityScheme"].Value.Type)
			require.NotNil(t, swagger.Components.SecuritySchemes["TestSecurityScheme"].Value.Scheme)
			require.Equal(t, "http", swagger.Components.SecuritySchemes["TestSecurityScheme"].Value.Type)
			require.Equal(t, "basic", swagger.Components.SecuritySchemes["TestSecurityScheme"].Value.Scheme)
		},
	},
	{
		name:            "PathRef",
		contentTemplate: relativePathDocsRefTemplate,
		testFunc: func(t *testing.T, swagger *Swagger) {
			require.NotNil(t, swagger.Paths["/pets"])
			require.NotNil(t, swagger.Paths["/pets"].Get.Responses["200"])
			require.NotNil(t, swagger.Paths["/pets"].Get.Responses["200"].Value.Content["application/json"])
		},
	},
}

func TestLoadSpecWithRelativeDocumentRefs(t *testing.T) {
	for _, td := range relativeDocRefsTestDataEntries {
		t.Logf("testcase '%s'", td.name)

		spec := []byte(td.contentTemplate)
		loader := NewSwaggerLoader()
		loader.IsExternalRefsAllowed = true
		swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testdata/"})
		require.NoError(t, err)
		td.testFunc(t, swagger)
	}
}

const relativeSchemaDocsRefTemplate = `
openapi: 3.0.0
info: 
  title: ""
  version: "1.0"
paths: {}
components: 
  schemas: 
    TestSchema: 
      $ref: relativeDocs/CustomTestSchema.yml
`

const relativeResponseDocsRefTemplate = `
openapi: 3.0.0
info: 
  title: ""
  version: "1.0"
paths: {}
components: 
  responses: 
    TestResponse: 
      $ref: relativeDocs/CustomTestResponse.yml
`

const relativeParameterDocsRefTemplate = `
openapi: 3.0.0
info:
  title: ""
  version: "1.0"
paths: {}
components:
  parameters:
    TestParameter: 
      $ref: relativeDocs/CustomTestParameter.yml
`

const relativeExampleDocsRefTemplate = `
openapi: 3.0.0
info:
  title: ""
  version: "1.0"
paths: {}
components:
  examples:
    TestExample:
      $ref: relativeDocs/CustomTestExample.yml
`

const relativeRequestDocsRefTemplate = `
openapi: 3.0.0
info:
  title: ""
  version: "1.0"
paths: {}
components:
  requestBodies:
    TestRequestBody:
      $ref: relativeDocs/CustomTestRequestBody.yml
`

const relativeHeaderDocsRefTemplate = `
openapi: 3.0.0
info:
  title: ""
  version: "1.0"
paths: {}
components:
  headers:
    TestHeader:
      $ref: relativeDocs/CustomTestHeader.yml
`

const relativeSecuritySchemeDocsRefTemplate = `
openapi: 3.0.0
info:
  title: ""
  version: "1.0"
paths: {}
components:
  securitySchemes:
    TestSecurityScheme:
      $ref: relativeDocs/CustomTestSecurityScheme.yml
`
const relativePathDocsRefTemplate = `
openapi: 3.0.0
info:
  title: ""
  version: "2.0"
paths:
  /pets:
    $ref: relativeDocs/CustomTestPath.yml
`

func TestLoadSpecWithRelativeDocumentRefs2(t *testing.T) {
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromFile("testdata/relativeDocsUseDocumentPath/openapi/openapi.yml")

	require.NoError(t, err)

	// path in nested directory
	// check parameter
	nestedDirPath := swagger.Paths["/pets/{id}"]
	require.Equal(t, "param", nestedDirPath.Patch.Parameters[0].Value.Name)
	require.Equal(t, "path", nestedDirPath.Patch.Parameters[0].Value.In)
	require.Equal(t, true, nestedDirPath.Patch.Parameters[0].Value.Required)

	// check header
	require.Equal(t, "header", nestedDirPath.Patch.Responses["200"].Value.Headers["X-Rate-Limit-Reset"].Value.Description)

	// check request body
	require.Equal(t, "example request", nestedDirPath.Patch.RequestBody.Value.Description)

	// check response schema and example
	require.Equal(t, nestedDirPath.Patch.Responses["200"].Value.Content["application/json"].Schema.Value.Type, "string")
	expectedExample := "hello"
	require.Equal(t, expectedExample, nestedDirPath.Patch.Responses["200"].Value.Content["application/json"].Examples["CustomTestExample"].Value.Value)

	// path in more nested directory
	// check parameter
	moreNestedDirPath := swagger.Paths["/pets/{id}/{city}"]
	require.Equal(t, "param", moreNestedDirPath.Patch.Parameters[0].Value.Name)
	require.Equal(t, "path", moreNestedDirPath.Patch.Parameters[0].Value.In)
	require.Equal(t, true, moreNestedDirPath.Patch.Parameters[0].Value.Required)

	// check header
	require.Equal(t, "header", nestedDirPath.Patch.Responses["200"].Value.Headers["X-Rate-Limit-Reset"].Value.Description)

	// check request body
	require.Equal(t, "example request", moreNestedDirPath.Patch.RequestBody.Value.Description)

	// check response schema and example
	require.Equal(t, "string", moreNestedDirPath.Patch.Responses["200"].Value.Content["application/json"].Schema.Value.Type)
	require.Equal(t, moreNestedDirPath.Patch.Responses["200"].Value.Content["application/json"].Examples["CustomTestExample"].Value.Value, expectedExample)
}
