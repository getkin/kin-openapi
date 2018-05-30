package openapi3_test

import (
	"encoding/base64"
	"github.com/ronniedada/kin-openapi/openapi3"
	"github.com/ronniedada/kin-test/jsontest"
	"testing"
)

type SchemaExample struct {
	Schema        *openapi3.Schema
	Serialization interface{}
	AllValid      []interface{}
	AllInvalid    []interface{}
}

var SchemaExamples = []SchemaExample{
	// EMPTY SCHEMA
	{
		Schema: &openapi3.Schema{},
		AllValid: []interface{}{
			nil,
			false,
			true,
			3.14,
			"",
			[]interface{}{},
			map[string]interface{}{},
		},
	},

	// BOOLEAN
	{
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

	// NUMBER
	{
		Schema: openapi3.NewFloat64Schema().
			WithMin(2.5).
			WithMax(3.5),
		Serialization: map[string]interface{}{
			"type": "number",
			"min":  2.5,
			"max":  3.5,
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
			// NaN and Inf don't even parse so they are not here
			"",
			[]interface{}{},
			map[string]interface{}{},
		},
	},

	// STRING
	{
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

	// STRING: format "date-time"
	{
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
		},
		AllInvalid: []interface{}{
			nil,
			3.14,
			"2017-12-31",
			"2017-12-31T11:59:59\n",
		},
	},

	// STRING: format "date-time"
	{
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

	// ARRAY
	{
		Schema: &openapi3.Schema{
			Type:     "array",
			MinItems: 2,
			MaxItems: openapi3.Int64Ptr(3),
			Items:    openapi3.NewFloat64Schema().NewRef(),
		},
		Serialization: map[string]interface{}{
			"type":     "array",
			"minItems": 2,
			"maxItems": 3,
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
				1, 2, 3, 4,
			},
		},
	},

	// OBJECT
	{
		Schema: &openapi3.Schema{
			Type: "object",
			Properties: map[string]*openapi3.SchemaRef{
				"numberProperty": openapi3.NewFloat64Schema().NewRef(),
			},
		},
		Serialization: map[string]interface{}{
			"type": "object",
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
				"otherPropery":   nil,
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
			Type: "object",
			AdditionalPropertiesAllowed: true,
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

	// NOT
	{
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

	// ANY OF
	{
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
					"type": "number",
					"min":  1,
					"max":  2,
				},
				map[string]interface{}{
					"type": "number",
					"min":  2,
					"max":  3,
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

	// ALL OF
	{
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
					"type": "number",
					"min":  1,
					"max":  2,
				},
				map[string]interface{}{
					"type": "number",
					"min":  2,
					"max":  3,
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

	// ONE OF
	{
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
					"type": "number",
					"min":  1,
					"max":  2,
				},
				map[string]interface{}{
					"type": "number",
					"min":  2,
					"max":  3,
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

func TestSchemas(t *testing.T) {
	for _, example := range SchemaExamples {
		schema := example.Schema
		if value := example.Serialization; value != nil {
			jsontest.Expect(t, schema).Value(value)
			jsontest.ExpectMarshallingAndUnmarshalling(t, value, &openapi3.Schema{}).Value(value)
		}
		for _, value := range example.AllValid {
			jsontest.Expect(t, value).MatchesSchema(schema)
		}
		for _, value := range example.AllInvalid {
			jsontest.Expect(t, value).DoesNotMatchSchema(schema)
		}
	}
}
