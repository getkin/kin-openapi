package openapi3_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

type schemaExample struct {
	Title         string
	Schema        *openapi3.Schema
	Serialization interface{}
	AllValid      []interface{}
	AllInvalid    []interface{}
}

func TestSchemas(t *testing.T) {
	openapi3.DefineStringFormat("uuid", openapi3.FormatOfStringForUUIDOfRFC4122)
	for _, example := range schemaExamples {
		t.Run(example.Title, testSchema(t, example))
	}
}

func testSchema(t *testing.T, example schemaExample) func(*testing.T) {
	return func(t *testing.T) {
		schema := example.Schema
		if serialized := example.Serialization; serialized != nil {
			jsonSerialized, err := json.Marshal(serialized)
			require.NoError(t, err)
			jsonSchema, err := json.Marshal(schema)
			require.NoError(t, err)
			require.JSONEq(t, string(jsonSerialized), string(jsonSchema))
			var dataUnserialized openapi3.Schema
			err = json.Unmarshal(jsonSerialized, &dataUnserialized)
			require.NoError(t, err)
			var dataSchema openapi3.Schema
			err = json.Unmarshal(jsonSchema, &dataSchema)
			require.NoError(t, err)
			require.Equal(t, dataUnserialized, dataSchema)
		}
		for _, value := range example.AllValid {
			err := validateSchema(t, schema, value)
			require.NoError(t, err)
		}
		for _, value := range example.AllInvalid {
			err := validateSchema(t, schema, value)
			require.Error(t, err)
		}
		// NaN and Inf aren't valid JSON but are handled
		for _, value := range []interface{}{math.NaN(), math.Inf(-1), math.Inf(+1)} {
			err := schema.VisitJSON(value)
			require.Error(t, err)
		}
	}
}

func validateSchema(t *testing.T, schema *openapi3.Schema, value interface{}) error {
	data, err := json.Marshal(value)
	require.NoError(t, err)
	var val interface{}
	err = json.Unmarshal(data, &val)
	require.NoError(t, err)
	return schema.VisitJSON(val)
}

var schemaExamples = []schemaExample{
	{
		Title:         "EMPTY SCHEMA",
		Schema:        &openapi3.Schema{},
		Serialization: map[string]interface{}{
			// This OA3 schema is exactly this draft-04 schema:
			//   {"not": {"type": "null"}}
		},
		AllValid: []interface{}{
			false,
			true,
			3.14,
			"",
			[]interface{}{},
			map[string]interface{}{},
		},
		AllInvalid: []interface{}{
			nil,
		},
	},

	{
		Title:  "JUST NULLABLE",
		Schema: openapi3.NewSchema().WithNullable(),
		Serialization: map[string]interface{}{
			// This OA3 schema is exactly both this draft-04 schema: {} and:
			// {anyOf: [type:string, type:number, type:integer, type:boolean
			//         ,{type:array, items:{}}, type:object]}
			"nullable": true,
		},
		AllValid: []interface{}{
			nil,
			false,
			true,
			0,
			0.0,
			3.14,
			"",
			[]interface{}{},
			map[string]interface{}{},
		},
	},

	{
		Title:  "NULLABLE BOOLEAN",
		Schema: openapi3.NewBoolSchema().WithNullable(),
		Serialization: map[string]interface{}{
			"nullable": true,
			"type":     "boolean",
		},
		AllValid: []interface{}{
			nil,
			false,
			true,
		},
		AllInvalid: []interface{}{
			0,
			0.0,
			3.14,
			"",
			[]interface{}{},
			map[string]interface{}{},
		},
	},

	{
		Title:  "BOOLEAN",
		Schema: openapi3.NewBoolSchema(),
		Serialization: map[string]interface{}{
			"type": "boolean",
		},
		AllValid: []interface{}{
			false,
			true,
		},
		AllInvalid: []interface{}{
			nil,
			3.14,
			"",
			[]interface{}{},
			map[string]interface{}{},
		},
	},

	{
		Title: "NUMBER",
		Schema: openapi3.NewFloat64Schema().
			WithMin(2.5).
			WithMax(3.5),
		Serialization: map[string]interface{}{
			"type":    "number",
			"minimum": 2.5,
			"maximum": 3.5,
		},
		AllValid: []interface{}{
			2.5,
			3.14,
			3.5,
		},
		AllInvalid: []interface{}{
			nil,
			false,
			true,
			2.4,
			3.6,
			"",
			[]interface{}{},
			map[string]interface{}{},
		},
	},

	{
		Title: "INTEGER",
		Schema: openapi3.NewInt64Schema().
			WithMin(2).
			WithMax(5),
		Serialization: map[string]interface{}{
			"type":    "integer",
			"format":  "int64",
			"minimum": 2,
			"maximum": 5,
		},
		AllValid: []interface{}{
			2,
			5,
		},
		AllInvalid: []interface{}{
			nil,
			false,
			true,
			1,
			6,
			3.5,
			"",
			[]interface{}{},
			map[string]interface{}{},
		},
	},

	{
		Title: "STRING",
		Schema: openapi3.NewStringSchema().
			WithMinLength(2).
			WithMaxLength(3).
			WithPattern("^[abc]+$"),
		Serialization: map[string]interface{}{
			"type":      "string",
			"minLength": 2,
			"maxLength": 3,
			"pattern":   "^[abc]+$",
		},
		AllValid: []interface{}{
			"ab",
			"abc",
		},
		AllInvalid: []interface{}{
			nil,
			false,
			true,
			3.14,
			"a",
			"xy",
			"aaaa",
			[]interface{}{},
			map[string]interface{}{},
		},
	},

	{
		Title:  "STRING: optional format 'uuid'",
		Schema: openapi3.NewUUIDSchema(),
		Serialization: map[string]interface{}{
			"type":   "string",
			"format": "uuid",
		},
		AllValid: []interface{}{
			"dd7d8481-81a3-407f-95f0-a2f1cb382a4b",
			"dcba3901-2fba-48c1-9db2-00422055804e",
			"ace8e3be-c254-4c10-8859-1401d9a9d52a",
		},
		AllInvalid: []interface{}{
			nil,
			"g39840b1-d0ef-446d-e555-48fcca50a90a",
			"4cf3i040-ea14-4daa-b0b5-ea9329473519",
			"aaf85740-7e27-4b4f-b4554-a03a43b1f5e3",
			"56f5bff4-z4b6-48e6-a10d-b6cf66a83b04",
		},
	},

	{
		Title:  "STRING: format 'date-time'",
		Schema: openapi3.NewDateTimeSchema(),
		Serialization: map[string]interface{}{
			"type":   "string",
			"format": "date-time",
		},
		AllValid: []interface{}{
			"2017-12-31T11:59:59",
			"2017-12-31T11:59:59Z",
			"2017-12-31T11:59:59-11:30",
			"2017-12-31T11:59:59+11:30",
			"2017-12-31T11:59:59.999+11:30",
			"2017-12-31T11:59:59.999Z",
		},
		AllInvalid: []interface{}{
			nil,
			3.14,
			"2017-12-31",
			"2017-12-31T11:59:59\n",
			"2017-12-31T11:59:59.+11:30",
			"2017-12-31T11:59:59.Z",
		},
	},

	{
		Title:  "STRING: format 'date-time'",
		Schema: openapi3.NewBytesSchema(),
		Serialization: map[string]interface{}{
			"type":   "string",
			"format": "byte",
		},
		AllValid: []interface{}{
			"",
			base64.StdEncoding.EncodeToString(func() []byte {
				data := make([]byte, 1024)
				for i := range data {
					data[i] = byte(i)
				}
				return data
			}()),
			base64.URLEncoding.EncodeToString(func() []byte {
				data := make([]byte, 1024)
				for i := range data {
					data[i] = byte(i)
				}
				return data
			}()),
		},
		AllInvalid: []interface{}{
			nil,
			" ",
			"\n",
			"%",
		},
	},

	{
		Title: "ARRAY",
		Schema: &openapi3.Schema{
			Type:        "array",
			MinItems:    2,
			MaxItems:    openapi3.Uint64Ptr(3),
			UniqueItems: true,
			Items:       openapi3.NewFloat64Schema().NewRef(),
		},
		Serialization: map[string]interface{}{
			"type":        "array",
			"minItems":    2,
			"maxItems":    3,
			"uniqueItems": true,
			"items": map[string]interface{}{
				"type": "number",
			},
		},
		AllValid: []interface{}{
			[]interface{}{
				1, 2,
			},
			[]interface{}{
				1, 2, 3,
			},
		},
		AllInvalid: []interface{}{
			nil,
			3.14,
			[]interface{}{
				1,
			},
			[]interface{}{
				42, 42,
			},
			[]interface{}{
				1, 2, 3, 4,
			},
		},
	},
	{
		Title: "ARRAY : items format 'object'",
		Schema: &openapi3.Schema{
			Type:        "array",
			UniqueItems: true,
			Items: (&openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"key1": openapi3.NewFloat64Schema().NewRef(),
				},
			}).NewRef(),
		},
		Serialization: map[string]interface{}{
			"type":        "array",
			"uniqueItems": true,
			"items": map[string]interface{}{
				"properties": map[string]interface{}{
					"key1": map[string]interface{}{
						"type": "number",
					},
				},
				"type": "object",
			},
		},
		AllValid: []interface{}{
			[]interface{}{
				map[string]interface{}{
					"key1": 1,
					"key2": 1,
					// Additioanl properties will make object different
					// By default additionalProperties is true
				},
				map[string]interface{}{
					"key1": 1,
				},
			},
			[]interface{}{
				map[string]interface{}{
					"key1": 1,
				},
				map[string]interface{}{
					"key1": 2,
				},
			},
		},
		AllInvalid: []interface{}{
			[]interface{}{
				map[string]interface{}{
					"key1": 1,
				},
				map[string]interface{}{
					"key1": 1,
				},
			},
		},
	},

	{
		Title: "ARRAY : items format 'object' and object with a property of array type ",
		Schema: &openapi3.Schema{
			Type:        "array",
			UniqueItems: true,
			Items: (&openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"key1": (&openapi3.Schema{
						Type:        "array",
						UniqueItems: true,
						Items:       openapi3.NewFloat64Schema().NewRef(),
					}).NewRef(),
				},
			}).NewRef(),
		},
		Serialization: map[string]interface{}{
			"type":        "array",
			"uniqueItems": true,
			"items": map[string]interface{}{
				"properties": map[string]interface{}{
					"key1": map[string]interface{}{
						"type":        "array",
						"uniqueItems": true,
						"items": map[string]interface{}{
							"type": "number",
						},
					},
				},
				"type": "object",
			},
		},
		AllValid: []interface{}{
			[]interface{}{
				map[string]interface{}{
					"key1": []interface{}{
						1, 2,
					},
				},
				map[string]interface{}{
					"key1": []interface{}{
						3, 4,
					},
				},
			},
			[]interface{}{ // Slice have items with the same value but with different index will treated as different slices
				map[string]interface{}{
					"key1": []interface{}{
						10, 9,
					},
				},
				map[string]interface{}{
					"key1": []interface{}{
						9, 10,
					},
				},
			},
		},
		AllInvalid: []interface{}{
			[]interface{}{ // Violate outer array uniqueItems: true
				map[string]interface{}{
					"key1": []interface{}{
						9, 9,
					},
				},
				map[string]interface{}{
					"key1": []interface{}{
						9, 9,
					},
				},
			},
			[]interface{}{ // Violate inner(array in object) array uniqueItems: true
				map[string]interface{}{
					"key1": []interface{}{
						9, 9,
					},
				},
				map[string]interface{}{
					"key1": []interface{}{
						8, 8,
					},
				},
			},
		},
	},

	{
		Title: "ARRAY : items format 'array'",
		Schema: &openapi3.Schema{
			Type:        "array",
			UniqueItems: true,
			Items: (&openapi3.Schema{
				Type:        "array",
				UniqueItems: true,
				Items:       openapi3.NewFloat64Schema().NewRef(),
			}).NewRef(),
		},
		Serialization: map[string]interface{}{
			"type":        "array",
			"uniqueItems": true,
			"items": map[string]interface{}{
				"items": map[string]interface{}{
					"type": "number",
				},
				"uniqueItems": true,
				"type":        "array",
			},
		},
		AllValid: []interface{}{
			[]interface{}{
				[]interface{}{1, 2},
				[]interface{}{3, 4},
			},
			[]interface{}{ // Slice have items with the same value but with different index will treated as different slices
				[]interface{}{1, 2},
				[]interface{}{2, 1},
			},
		},
		AllInvalid: []interface{}{
			[]interface{}{ // Violate outer array uniqueItems: true
				[]interface{}{8, 9},
				[]interface{}{8, 9},
			},
			[]interface{}{ // Violate inner array uniqueItems: true
				[]interface{}{9, 9},
				[]interface{}{8, 8},
			},
		},
	},

	{
		Title: "ARRAY : items format 'array' and array with object type items",
		Schema: &openapi3.Schema{
			Type:        "array",
			UniqueItems: true,
			Items: (&openapi3.Schema{
				Type:        "array",
				UniqueItems: true,
				Items: (&openapi3.Schema{
					Type: "object",
					Properties: map[string]*openapi3.SchemaRef{
						"key1": openapi3.NewFloat64Schema().NewRef(),
					},
				}).NewRef(),
			}).NewRef(),
		},
		Serialization: map[string]interface{}{
			"type":        "array",
			"uniqueItems": true,
			"items": map[string]interface{}{
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"key1": map[string]interface{}{
							"type": "number",
						},
					},
				},
				"uniqueItems": true,
				"type":        "array",
			},
		},
		AllValid: []interface{}{
			[]interface{}{
				[]interface{}{
					map[string]interface{}{
						"key1": 1,
					},
				},
				[]interface{}{
					map[string]interface{}{
						"key1": 2,
					},
				},
			},
			[]interface{}{ // Slice have items with the same value but with different index will treated as different slices
				[]interface{}{
					map[string]interface{}{
						"key1": 1,
					},
					map[string]interface{}{
						"key1": 2,
					},
				},
				[]interface{}{
					map[string]interface{}{
						"key1": 2,
					},
					map[string]interface{}{
						"key1": 1,
					},
				},
			},
		},
		AllInvalid: []interface{}{
			[]interface{}{ // Violate outer array uniqueItems: true
				[]interface{}{
					map[string]interface{}{
						"key1": 1,
					},
					map[string]interface{}{
						"key1": 2,
					},
				},
				[]interface{}{
					map[string]interface{}{
						"key1": 1,
					},
					map[string]interface{}{
						"key1": 2,
					},
				},
			},
			[]interface{}{ // Violate inner array uniqueItems: true
				[]interface{}{
					map[string]interface{}{
						"key1": 1,
					},
					map[string]interface{}{
						"key1": 1,
					},
				},
				[]interface{}{
					map[string]interface{}{
						"key1": 2,
					},
					map[string]interface{}{
						"key1": 2,
					},
				},
			},
		},
	},

	{
		Title: "OBJECT",
		Schema: &openapi3.Schema{
			Type:     "object",
			MaxProps: openapi3.Uint64Ptr(2),
			Properties: map[string]*openapi3.SchemaRef{
				"numberProperty": openapi3.NewFloat64Schema().NewRef(),
			},
		},
		Serialization: map[string]interface{}{
			"type":          "object",
			"maxProperties": 2,
			"properties": map[string]interface{}{
				"numberProperty": map[string]interface{}{
					"type": "number",
				},
			},
		},
		AllValid: []interface{}{
			map[string]interface{}{},
			map[string]interface{}{
				"numberProperty": 3.14,
			},
			map[string]interface{}{
				"numberProperty": 3.14,
				"some prop":      nil,
			},
		},
		AllInvalid: []interface{}{
			nil,
			false,
			true,
			3.14,
			"",
			[]interface{}{},
			map[string]interface{}{
				"numberProperty": "abc",
			},
			map[string]interface{}{
				"numberProperty": 3.14,
				"some prop":      42,
				"third":          "prop",
			},
		},
	},
	{
		Schema: &openapi3.Schema{
			Type: "object",
			AdditionalProperties: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: "number",
				},
			},
		},
		Serialization: map[string]interface{}{
			"type": "object",
			"additionalProperties": map[string]interface{}{
				"type": "number",
			},
		},
		AllValid: []interface{}{
			map[string]interface{}{},
			map[string]interface{}{
				"x": 3.14,
				"y": 3.14,
			},
		},
		AllInvalid: []interface{}{
			map[string]interface{}{
				"x": "abc",
			},
		},
	},
	{
		Schema: &openapi3.Schema{
			Type:                        "object",
			AdditionalPropertiesAllowed: openapi3.BoolPtr(true),
		},
		Serialization: map[string]interface{}{
			"type":                 "object",
			"additionalProperties": true,
		},
		AllValid: []interface{}{
			map[string]interface{}{},
			map[string]interface{}{
				"x": false,
				"y": 3.14,
			},
		},
	},

	{
		Title: "NOT",
		Schema: &openapi3.Schema{
			Not: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Enum: []interface{}{
						nil,
						true,
						3.14,
						"not this",
					},
				},
			},
		},
		Serialization: map[string]interface{}{
			"not": map[string]interface{}{
				"enum": []interface{}{
					nil,
					true,
					3.14,
					"not this",
				},
			},
		},
		AllValid: []interface{}{
			false,
			2,
			"abc",
		},
		AllInvalid: []interface{}{
			nil,
			true,
			3.14,
			"not this",
		},
	},

	{
		Title: "ANY OF",
		Schema: &openapi3.Schema{
			AnyOf: []*openapi3.SchemaRef{
				{
					Value: openapi3.NewFloat64Schema().
						WithMin(1).
						WithMax(2),
				},
				{
					Value: openapi3.NewFloat64Schema().
						WithMin(2).
						WithMax(3),
				},
			},
		},
		Serialization: map[string]interface{}{
			"anyOf": []interface{}{
				map[string]interface{}{
					"type":    "number",
					"minimum": 1,
					"maximum": 2,
				},
				map[string]interface{}{
					"type":    "number",
					"minimum": 2,
					"maximum": 3,
				},
			},
		},
		AllValid: []interface{}{
			1,
			2,
			3,
		},
		AllInvalid: []interface{}{
			0,
			4,
		},
	},

	{
		Title: "ALL OF",
		Schema: &openapi3.Schema{
			AllOf: []*openapi3.SchemaRef{
				{
					Value: openapi3.NewFloat64Schema().
						WithMin(1).
						WithMax(2),
				},
				{
					Value: openapi3.NewFloat64Schema().
						WithMin(2).
						WithMax(3),
				},
			},
		},
		Serialization: map[string]interface{}{
			"allOf": []interface{}{
				map[string]interface{}{
					"type":    "number",
					"minimum": 1,
					"maximum": 2,
				},
				map[string]interface{}{
					"type":    "number",
					"minimum": 2,
					"maximum": 3,
				},
			},
		},
		AllValid: []interface{}{
			2,
		},
		AllInvalid: []interface{}{
			0,
			1,
			3,
			4,
		},
	},

	{
		Title: "ONE OF",
		Schema: &openapi3.Schema{
			OneOf: []*openapi3.SchemaRef{
				{
					Value: openapi3.NewFloat64Schema().
						WithMin(1).
						WithMax(2),
				},
				{
					Value: openapi3.NewFloat64Schema().
						WithMin(2).
						WithMax(3),
				},
			},
		},
		Serialization: map[string]interface{}{
			"oneOf": []interface{}{
				map[string]interface{}{
					"type":    "number",
					"minimum": 1,
					"maximum": 2,
				},
				map[string]interface{}{
					"type":    "number",
					"minimum": 2,
					"maximum": 3,
				},
			},
		},
		AllValid: []interface{}{
			1,
			3,
		},
		AllInvalid: []interface{}{
			0,
			2,
			4,
		},
	},
}

type schemaTypeExample struct {
	Title      string
	Schema     *openapi3.Schema
	AllValid   []string
	AllInvalid []string
}

func TestTypes(t *testing.T) {
	for _, example := range typeExamples {
		t.Run(example.Title, testType(t, example))
	}
}

func testType(t *testing.T, example schemaTypeExample) func(*testing.T) {
	return func(t *testing.T) {
		baseSchema := example.Schema
		for _, typ := range example.AllValid {
			schema := baseSchema.WithFormat(typ)
			err := schema.Validate(context.TODO())
			require.NoError(t, err)
		}
		for _, typ := range example.AllInvalid {
			schema := baseSchema.WithFormat(typ)
			err := schema.Validate(context.TODO())
			require.Error(t, err)
		}
	}
}

var typeExamples = []schemaTypeExample{
	{
		Title:  "STRING",
		Schema: openapi3.NewStringSchema(),
		AllValid: []string{
			"",
			"byte",
			"binary",
			"date",
			"date-time",
			"password",
			// Not supported but allowed:
			"uri",
		},
		AllInvalid: []string{
			"code/golang",
		},
	},

	{
		Title:  "NUMBER",
		Schema: openapi3.NewFloat64Schema(),
		AllValid: []string{
			"",
			"float",
			"double",
		},
		AllInvalid: []string{
			"f32",
		},
	},

	{
		Title:  "INTEGER",
		Schema: openapi3.NewIntegerSchema(),
		AllValid: []string{
			"",
			"int32",
			"int64",
		},
		AllInvalid: []string{
			"uint8",
		},
	},
}

func TestSchemaErrors(t *testing.T) {
	for _, example := range schemaErrorExamples {
		t.Run(example.Title, testSchemaError(t, example))
	}
}

func testSchemaError(t *testing.T, example schemaErrorExample) func(*testing.T) {
	return func(t *testing.T) {
		msg := example.Error.Error()
		require.True(t, strings.Contains(msg, example.Want))
	}
}

type schemaErrorExample struct {
	Title string
	Error *openapi3.SchemaError
	Want  string
}

var schemaErrorExamples = []schemaErrorExample{
	{
		Title: "SIMPLE",
		Error: &openapi3.SchemaError{
			Value:  1,
			Schema: &openapi3.Schema{},
			Reason: "SIMPLE",
		},
		Want: "SIMPLE",
	},
	{
		Title: "NEST",
		Error: &openapi3.SchemaError{
			Value:  1,
			Schema: &openapi3.Schema{},
			Reason: "PARENT",
			Origin: &openapi3.SchemaError{
				Value:  1,
				Schema: &openapi3.Schema{},
				Reason: "NEST",
			},
		},
		Want: "NEST",
	},
}

func TestRegisterArrayUniqueItemsChecker(t *testing.T) {
	var (
		checker = func(items []interface{}) bool {
			return false
		}
		scheme = openapi3.Schema{
			Type:        "array",
			UniqueItems: true,
			Items:       openapi3.NewStringSchema().NewRef(),
		}
		val = []interface{}{"1", "2", "3"}
		err error
	)

	// Fist checked by predefined function
	err = scheme.VisitJSON(val)
	require.NoError(t, err)

	// Register a function will always return false when check if a
	// slice has unique items, then use a slice indeed has unique
	// items to verify that check unique items will failed.
	openapi3.RegisterArrayUniqueItemsChecker(checker)

	err = scheme.VisitJSON(val)
	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "Duplicate items found"))
}
