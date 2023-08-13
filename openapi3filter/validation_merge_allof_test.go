package openapi3filter

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"

	"github.com/stretchr/testify/require"
)

type Test struct {
	data    []byte
	wantErr bool
}

// validate that merge is successful when additionalProperties is a Schema
func TestMerge_AdditionalProperties_Schema(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate range
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    age:
                      type: integer
                    name:
                      type: string
                  additionalProperties:
                    type: integer
                - type: object
                  properties:
                    height:
                      type: integer
                  additionalProperties:
                    maximum: 100
      responses:
        '200':
          description: Ok
`

	tests := []Test{
		{
			[]byte(`{"age": 1}`),
			false,
		},
		{
			[]byte(`{"height": 1000}`),
			false,
		},
		{
			[]byte(`{"additionalProp": 1}`),
			false,
		},
		{
			[]byte(`{"additionalProp": 101}`),
			true,
		},
	}
	validateConsistency(t, spec, tests)
}

// validate only intersecting properties if one of the additionalProperties is false
func TestMerge_AdditionalProperties_False(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate range
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    age:
                      type: integer
                    name:
                      type: string
                  additionalProperties: true
                - type: object
                  properties:
                    height:
                      type: integer
                  additionalProperties: false
      responses:
        '200':
          description: Ok
`
	tests := []Test{
		{
			[]byte(`{"height": 1}`),
			false,
		},
		{
			[]byte(`{"height": "1"}`),
			true,
		},
		{
			[]byte(`{"name": "a", "age:": 1, "height": 1}`),
			true,
		},
	}
	validateConsistency(t, spec, tests)

	spec2 := `
   openapi: 3.0.0
   info:
     title: Example integer enum
     version: '0.1'
   paths:
     /sample:
       put:
         requestBody:
           required: true
           content:
             application/json:
               schema:
                 allOf:
                   - type: object
                     properties:
                       test:
                         enum: ["1", "5", "3"]
                       prop1:
                         type: string
                     additionalProperties: true
                   - type: object
                     properties:
                       test1:
                         enum: ["1", "8", "7"]
                     additionalProperties: false
                   - type: object
                     properties:
                       test1:
                         enum: ["3", "8", "5"]
                     additionalProperties: false					 
         responses:
           '200':
             description: Ok
   `

	tests = []Test{
		{
			[]byte(`{"test1": "8"}`),
			false,
		},
		{
			[]byte(`{"test1": "1"}`),
			true,
		},
		{
			[]byte(`{"prop1": "string"}`),
			true,
		},
	}

	validateConsistency(t, spec2, tests)
}

// non-conflicting Properties range can be merged
func TestMergePropertiesRange(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate items range is restrictive
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  minProperties: 1
                  maxProperties: 3
                - type: object
                  minProperties: 2
                  maxProperties: 4
      responses:
        '200':
          description: Ok
`
	tests := []Test{
		{
			[]byte(`{"a": 1, "b": 2}`),
			false,
		},
		{
			[]byte(`{"a": 1, "b": 2, "c": 3}`),
			false,
		},
		{
			[]byte(`{"a": 1, "b": 2, "c": 3, "d": 4}`),
			true,
		},
		{
			[]byte(`{"a": 1}`),
			true,
		},
		{
			[]byte(`{}`),
			true,
		},
	}

	validateConsistency(t, spec, tests)
}

// non-conflicting Items range can be merged
func TestMergeItemsRange(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate items range is restrictive
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    test:
                      type: array
                      items:
                        type: integer
                      minItems: 1
                      maxItems: 3
                - type: object
                  properties:
                    test:
                      type: array
                      items:
                        type: integer
                      minItems: 2
                      maxItems: 4
      responses:
        '200':
          description: Ok
`
	tests := []Test{
		{
			[]byte(`{"test": [1, 2]}`),
			false,
		},
		{
			[]byte(`{"test": [1, 2, 3]}`),
			false,
		},
		{
			[]byte(`{"test": [1, 2, 3, 4]}`),
			true,
		},
		{
			[]byte(`{"test": [1]}`),
			true,
		},
		{
			[]byte(`{"test": []}`),
			true,
		},
	}

	validateConsistency(t, spec, tests)
}

func TestMergeItems(t *testing.T) {
	//todo: cleanup

	const spec = `
openapi: 3.0.0
info:
  title: Validate items of type integer
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    test:
                      type: array
                      items:
                        type: integer
                - type: object
                  properties:
                    test:
                      type: array
                      items:
                        type: integer
      responses:
        '200':
          description: Ok
`
	tests := []Test{
		{
			[]byte(`{"test": [1, 2, 3]}`),
			false,
		},
		{
			[]byte(`{"test": ["abc"]}`),
			true,
		},
	}

	validateConsistency(t, spec, tests)

	const spec2 = `
openapi: 3.0.0
info:
  title: Validate items of objects 
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    test:
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            type: string
                - type: object
                  properties:
                    test:
                      type: array
                      items:
                        type: object
                        properties:
                          id:
                            type: integer
      responses:
        '200':
          description: Ok
`

	tests = []Test{
		{
			[]byte(`{"test": [{"id": 1, "name": "abc"}]}`),
			false,
		},
		{
			[]byte(`{"test": [{"id": "1"}]}`),
			true,
		},
	}

	validateConsistency(t, spec2, tests)
}

// conflicting uniqueItems can be merged
func TestMergeUniqueItems(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate merge of unique items
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    test:
                      type: array
                      items:
                        type: integer
                      uniqueItems: true
                - type: object
                  properties:
                    test:
                      type: array
                      items:
                        type: integer
                      uniqueItems: false
      responses:
        '200':
          description: Ok
`

	tests := []Test{
		{
			[]byte(`{"test": [1, 2, 3]}`),
			false,
		},
		{
			[]byte(`{"test": [1, 1]}`),
			true,
		},
	}
	validateConsistency(t, spec, tests)
}

// non-conflicting properties with required can be merged
func TestMergeRequired(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate range
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    name:
                      type: string
                    id:
                      type: integer
                  required:
                    - id
                - type: object
                  properties:
                    age:
                      type: integer
                    id:
                      type: integer
                  required:
                    - age
                    - id
                - type: object
                  properties:
                    nickname:
                      type: string
                  required:
                    - nickname
      responses:
        '200':
          description: Ok
`

	tests := []Test{
		{
			[]byte(`{"age": 1, "name": "abc", "id": 1, "nickname": "def"}`),
			false,
		},
		{
			[]byte(`{"age": 1, "name": "abc", "nickname": "def"}`),
			true,
		},
		{
			[]byte(`{"name": "abc", "id": 1, "nickname": "def"}`),
			true,
		},
		{
			[]byte(`{"age": "a", "name": "abc", "id": 1, "nickname": "def"}`),
			true,
		},
		{
			[]byte(`{"age": 1, "name": 100, "id": 1, "nickname": "def"}`),
			true,
		},
	}

	validateConsistency(t, spec, tests)
}

// multiple-of can always be merged
func TestMergeMultipleOf(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate multiple of
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    test:
                      type: integer
                      multipleOf: 12
                - type: object
                  properties:
                    test:
                      type: integer
                      multipleOf: 15
      responses:
        '200':
          description: Ok
`
	tests := []Test{
		{
			[]byte(`{"test": 61}`),
			true,
		},
		{
			[]byte(`{"test": 1}`),
			true,
		},
		{
			[]byte(`{"test": 60}`),
			false,
		},
		{
			[]byte(`{"test": 180}`),
			false,
		},
	}

	validateConsistency(t, spec, tests)
}

// minlength and maxlength can always be merged
func TestMergeStringRange(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate string range
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    test:
                      type: string
                      minLength: 1
                      maxLength: 10
                - type: object
                  properties:
                    test:
                      type: string
                      minLength: 5
                      maxLength: 9
      responses:
        '200':
          description: Ok
`

	tests := []Test{
		{
			[]byte(`{"test": "1234"}`),
			true,
		},
		{
			[]byte(`{"test": "12345678910"}`),
			true,
		},
		{
			[]byte(`{"test": "12345"}`),
			false,
		},
		{
			[]byte(`{"test": "123456789"}`),
			false,
		},
	}

	validateConsistency(t, spec, tests)
}

func TestMergeExclusiveRange(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate exclusive range
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    age:
                      type: integer
                      minimum: 10
                      maximum: 40
                      exclusiveMinimum: true
                      exclusiveMaximum: true
                - type: object
                  properties:
                    age:
                      type: integer
                      minimum: 5
                      maximum: 25
                      exclusiveMaximum: true
                      exclusiveMinimum: true
      responses:
        '200':
          description: Ok
`

	tests := []Test{
		{
			[]byte(`{"age": 10}`),
			true,
		},
		{
			[]byte(`{"age": 25}`),
			true,
		},
		{
			[]byte(`{"age": 11}`),
			false,
		},
		{
			[]byte(`{"age": 24}`),
			false,
		},
	}

	validateConsistency(t, spec, tests)
}

func TestMergeRange(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Validate range
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    age:
                      type: integer
                      minimum: 10
                      maximum: 40
                - type: object
                  properties:
                    age:
                      type: integer
                      minimum: 5
                      maximum: 25
      responses:
        '200':
          description: Ok
`

	tests := []Test{
		{
			[]byte(`{"age": 9}`),
			true,
		},
		{
			[]byte(`{"age": 26}`),
			true,
		},
		{
			[]byte(`{"age": 10}`),
			false,
		},
		{
			[]byte(`{"age": 25}`),
			false,
		},
	}

	validateConsistency(t, spec, tests)
}

// enum is merged as the intersection of all values
func TestMergeIntegerEnum(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Example integer enum
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - type: object
                  properties:
                    test1:
                      enum: ["1", "2", "3"]
                - type: object
                  properties:
                    test1:
                      enum: ["1", "2", "4"]
      responses:
        '200':
          description: Ok
`

	tests := []Test{
		{
			[]byte(`{"test1": "2"}`),
			false,
		},
		{
			[]byte(`{"test1": "1"}`),
			false,
		},
		{
			[]byte(`{"test1": "4"}`),
			true,
		},
		{
			[]byte(`{"test1": "3"}`),
			true,
		},
		{
			[]byte(`{"test1": ""}`),
			true,
		},
	}

	validateConsistency(t, spec, tests)
}

func TestMerge_AnyOf(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Example integer enum
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - anyOf:
                    - type: object
                      properties:
                        test1:
                          type: string
                        test2:
                          type: boolean
                    - type: object
                      properties:
                        test1:
                          type: number
                - anyOf:
                    - type: object
                      properties:
                        test2:
                          type: boolean
      responses:
        '200':
          description: Ok
`

	tests := []Test{
		{
			[]byte(`{"test1": 1}`),
			false,
		},
		{
			[]byte(`{"test2": 1}`),
			true,
		},
		{
			[]byte(`{"test2": true, "test1": 111}`),
			false,
		},
		{
			[]byte(`{"test2": 1}`),
			true,
		},
		{
			[]byte(`{"test3": 1}`),
			false,
		},
	}

	validateConsistency(t, spec, tests)
}

// Testing OneOf Merging Inside AllOf
func TestMerge_OneOf(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Testing OneOf Merging Inside AllOf
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                - oneOf:
                  - type: object
                    properties:
                      test1:
                        type: string
                      test2:
                        type: boolean
                - oneOf:
                  - type: object
                    properties:
                      test1:
                        type: string
      responses:
        '200':
          description: Ok
`
	tests := []Test{
		{
			[]byte(`{"test1": "string"}`),
			false,
		},
	}

	validateConsistency(t, spec, tests)
}

// testing Multiple `not` inside `allOf`
func TestMerge_Not(t *testing.T) {

	const spec = `
openapi: 3.0.0
info:
  title: Multiple 'not' inside 'allOf' Example
  version: '0.1'
paths:
  /sample:
    put:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              allOf:
                  - not:
                      type: object
                      properties:
                        test1:
                          type: string
                        test2:
                          type: object
                  - not:
                      type: object
                      properties:
                        test1:
                          type: number
                        test2:
                          type: boolean
      responses:
        '200':
          description: Ok
`
	tests := []Test{
		{
			[]byte(`{"test1": true}`),
			false,
		},
		{
			[]byte(`{"test2": "string"}`),
			false,
		},
		{
			[]byte(`{"test1": "string"}`),
			true,
		},
		{
			[]byte(`{"test2": true}`),
			true,
		},
	}

	validateConsistency(t, spec, tests)
}

func validateConsistency(t *testing.T, spec string, tests []Test) {
	nonMerged := runTests(t, spec, tests, false)
	merged := runTests(t, spec, tests, true)

	for i, test := range tests {
		if test.wantErr {
			require.Error(t, nonMerged[i])
			require.Error(t, merged[i])
		} else {
			require.NoError(t, nonMerged[i])
			require.NoError(t, merged[i])
		}
	}
}

func merge(doc *openapi3.T) *openapi3.T {
	schemaRef := doc.Paths.Find("/sample").Put.RequestBody.Value.Content.Get("application/json").Schema
	merged, err := openapi3.Merge(schemaRef.Value)
	if err != nil {
		log.Fatal(err.Error())
	}
	schemaRef.Value = merged
	return doc
}

func runTests(t *testing.T, spec string, tests []Test, shouldMerge bool) []error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))

	if shouldMerge {
		doc = merge(doc)
		// j, err := doc.MarshalJSON()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// PrettyPrintJSON(j)
		// log.Fatal("bye")
	}

	require.NoError(t, err)

	router, err := legacyrouter.NewRouter(doc)
	require.NoError(t, err)

	result := []error{}
	for _, tt := range tests {
		body := bytes.NewReader(tt.data)
		req, err := http.NewRequest("PUT", "/sample", body)
		require.NoError(t, err)
		req.Header.Add(headerCT, "application/json")

		route, pathParams, err := router.FindRoute(req)
		require.NoError(t, err)

		requestValidationInput := &RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		}
		err = ValidateRequest(loader.Context, requestValidationInput)
		result = append(result, err)

	}
	return result
}

func PrettyPrintJSON(rawJSON []byte) error {
	var data interface{}

	err := json.Unmarshal(rawJSON, &data)
	if err != nil {
		return err
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	log.Println(string(prettyJSON))
	return nil
}
