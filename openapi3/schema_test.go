package openapi3_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

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
		Schema: openapi3.NewUuidSchema(),
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		baseSchema := example.Schema
		for _, typ := range example.AllValid {
			schema := baseSchema.WithFormat(typ)
			err := schema.Validate(ctx)
			require.NoError(t, err)
		}
		for _, typ := range example.AllInvalid {
			schema := baseSchema.WithFormat(typ)
			err := schema.Validate(ctx)
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

func TestSchemaDefaultValues(t *testing.T) {
	for _, example := range defaultExamples {
		t.Run(example.Title, testDefaultValue(t, example))
	}
}

func testDefaultValue(t *testing.T, example schemaDefaultExample) func(*testing.T) {
	return func(t *testing.T) {
		baseSchema := example.Schema
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for ndx, arr := range [][]fmtDflt{example.AllValid, example.AllInvalid} {
			for _, typ := range arr {
				schema := baseSchema.WithDefault(typ.dflt)
				s, f := typ.fmt, schema.WithFormat
				if typ.pat != "" {
					s, f = typ.pat, schema.WithPattern
				}
				schema = f(s)

				err := schema.Validate(ctx)
				r := require.NoError
				if ndx > 0 {
					r = require.Error
				}
				r(t, err, typ)
			}
		}
	}
}

type fmtDflt struct {
	fmt  string
	pat  string
	dflt interface{}
}
type schemaDefaultExample struct {
	Title      string
	Schema     *openapi3.Schema
	AllValid   []fmtDflt
	AllInvalid []fmtDflt
}

var defaultExamples = []schemaDefaultExample{
	{
		Title:  "STRING",
		Schema: openapi3.NewStringSchema(),
		AllValid: []fmtDflt{
			{fmt: "byte", pat: "", dflt: "aGVsbG8sIHdvcmxk"},
			{fmt: "date", pat: "", dflt: "2019-10-01"},
			{fmt: "date-time", pat: "", dflt: "2019-10-01T00:11:22Z"},
			{fmt: "password", pat: "", dflt: "top-secret"}, // <-- this is not really validated other than it's a string
			{fmt: "", pat: "^1.+2$", dflt: "111112"},
			{fmt: "", pat: "^1.+2$", dflt: "1abc2"},
		},
		AllInvalid: []fmtDflt{
			{fmt: "byte", pat: "", dflt: "\r\t\n"},
			{fmt: "byte", pat: "", dflt: 123},
			{fmt: "date", pat: "", dflt: "19-10-01"},
			{fmt: "date", pat: "", dflt: "abcd"},
			{fmt: "date", pat: "", dflt: 123},
			{fmt: "date-time", pat: "", dflt: "66:11:22"},
			{fmt: "date-time", pat: "", dflt: "66:11:2"},
			{fmt: "date-time", pat: "", dflt: 123},
			{fmt: "password", pat: "", dflt: 123},
			{fmt: "", pat: "^1.+2$", dflt: 123},
		},
	},

	{
		Title:  "NUMBER",
		Schema: openapi3.NewFloat64Schema(),
		AllValid: []fmtDflt{
			{fmt: "", pat: "", dflt: 3.14159},
			{fmt: "float", pat: "", dflt: 3.14159},
			{fmt: "double", pat: "", dflt: 3.123123123123123123123},
		},
		AllInvalid: []fmtDflt{
			{fmt: "", pat: "", dflt: "abc"},
			{fmt: "float", pat: "", dflt: "abc"},
			{fmt: "float", pat: "", dflt: math.MaxFloat64},
			{fmt: "double", pat: "", dflt: "abc"},
			//{ fmt: "double", pat: "", dflt: 1E1200 },  // compiler catches that....
		},
	},

	{
		Title:  "INTEGER",
		Schema: openapi3.NewIntegerSchema(),
		AllValid: []fmtDflt{
			{fmt: "", pat: "", dflt: float64(12345)},
			{fmt: "int32", pat: "", dflt: float64(math.MinInt32)},
			{fmt: "int32", pat: "", dflt: float64(math.MaxInt32)},
			{fmt: "int64", pat: "", dflt: float64(-openapi3.MaxIntegerInFloat64Significand)},
			{fmt: "int64", pat: "", dflt: float64(openapi3.MaxIntegerInFloat64Significand)},
		},
		AllInvalid: []fmtDflt{
			{fmt: "", pat: "", dflt: "1233"},
			{fmt: "int32", pat: "", dflt: math.MinInt64},
			{fmt: "int32", pat: "", dflt: 3.14159},
			{fmt: "int64", pat: "", dflt: 3.14159},
			{fmt: "int64", pat: "", dflt: float64(openapi3.MaxIntegerInFloat64Significand + 3)},
			{fmt: "int64", pat: "", dflt: float64(-(openapi3.MaxIntegerInFloat64Significand + 3))},
		},
	},
	{
		Title:  "BOOL",
		Schema: openapi3.NewBoolSchema(),
		AllValid: []fmtDflt{
			{fmt: "", pat: "", dflt: "t"},
			{fmt: "", pat: "", dflt: "f"},
			{fmt: "", pat: "", dflt: "TRUE"},
			{fmt: "", pat: "", dflt: "true"},
			{fmt: "", pat: "", dflt: "FALSE"},
			{fmt: "", pat: "", dflt: "false"},
		},
		AllInvalid: []fmtDflt{
			{fmt: "", pat: "", dflt: "abc"},
			{fmt: "", pat: "", dflt: 12},
			{fmt: "", pat: "", dflt: 1.2},
		},
	},
}
