package openapi3

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	name             string
	schema           *Schema
	value            any
	extraNotContains []any
	options          []SchemaValidationOption
}

func TestIssue735(t *testing.T) {
	DefineStringFormatValidator("uuid", NewRegexpFormatValidator(FormatOfStringForUUIDOfRFC4122))
	DefineStringFormatValidator("email", NewRegexpFormatValidator(FormatOfStringForEmail))
	DefineIPv4Format()
	DefineIPv6Format()

	testCases := []testCase{
		{
			name:   "type string",
			schema: NewStringSchema(),
			value:  42,
		},
		{
			name:   "type boolean",
			schema: NewBoolSchema(),
			value:  42,
		},
		{
			name:   "type integer",
			schema: NewIntegerSchema(),
			value:  "foo",
		},
		{
			name:   "type number",
			schema: NewFloat64Schema(),
			value:  "foo",
		},
		{
			name:   "type array",
			schema: NewArraySchema(),
			value:  42,
		},
		{
			name:   "type object",
			schema: NewObjectSchema(),
			value:  42,
		},
		{
			name:   "min",
			schema: NewSchema().WithMin(100),
			value:  42,
		},
		{
			name:   "max",
			schema: NewSchema().WithMax(0),
			value:  42,
		},
		{
			name:   "exclusive min",
			schema: NewSchema().WithMin(100).WithExclusiveMin(true),
			value:  42,
		},
		{
			name:   "exclusive max",
			schema: NewSchema().WithMax(0).WithExclusiveMax(true),
			value:  42,
		},
		{
			name:   "multiple of",
			schema: &Schema{MultipleOf: Ptr(5.0)},
			value:  42,
		},
		{
			name:   "enum",
			schema: NewSchema().WithEnum(3, 5),
			value:  42,
		},
		{
			name:   "min length",
			schema: NewSchema().WithMinLength(100),
			value:  "foo",
		},
		{
			name:   "max length",
			schema: NewSchema().WithMaxLength(0),
			value:  "foo",
		},
		{
			name:   "pattern",
			schema: NewSchema().WithPattern("[0-9]"),
			value:  "foo",
		},
		{
			name:             "items",
			schema:           NewSchema().WithItems(NewStringSchema()),
			value:            []any{42},
			extraNotContains: []any{42},
		},
		{
			name:             "min items",
			schema:           NewSchema().WithMinItems(100),
			value:            []any{42},
			extraNotContains: []any{42},
		},
		{
			name:             "max items",
			schema:           NewSchema().WithMaxItems(0),
			value:            []any{42},
			extraNotContains: []any{42},
		},
		{
			name:             "unique items",
			schema:           NewSchema().WithUniqueItems(true),
			value:            []any{42, 42},
			extraNotContains: []any{42},
		},
		{
			name:             "min properties",
			schema:           NewSchema().WithMinProperties(100),
			value:            map[string]any{"foo": 42},
			extraNotContains: []any{42},
		},
		{
			name:             "max properties",
			schema:           NewSchema().WithMaxProperties(0),
			value:            map[string]any{"foo": 42},
			extraNotContains: []any{42},
		},
		{
			name:             "additional properties other schema type",
			schema:           NewSchema().WithAdditionalProperties(NewStringSchema()),
			value:            map[string]any{"foo": 42},
			extraNotContains: []any{42},
		},
		{
			name: "additional properties false",
			schema: &Schema{AdditionalProperties: AdditionalProperties{
				Has: Ptr(false),
			}},
			value:            map[string]any{"foo": 42},
			extraNotContains: []any{42},
		},
		{
			name: "invalid properties schema",
			schema: NewSchema().WithProperties(map[string]*Schema{
				"foo": NewStringSchema(),
			}),
			value:            map[string]any{"foo": 42},
			extraNotContains: []any{42},
		},
		// TODO: uncomment when https://github.com/getkin/kin-openapi/issues/502 is fixed
		//{
		//	name: "read only properties",
		//	schema: NewSchema().WithProperties(map[string]*Schema{
		//		"foo": {ReadOnly: true},
		//	}).WithoutAdditionalProperties(),
		//	value:            map[string]any{"foo": 42},
		//	extraNotContains: []any{42},
		//	options:          []SchemaValidationOption{VisitAsRequest()},
		//},
		//{
		//	name: "write only properties",
		//	schema: NewSchema().WithProperties(map[string]*Schema{
		//		"foo": {WriteOnly: true},
		//	}).WithoutAdditionalProperties(),
		//	value:            map[string]any{"foo": 42},
		//	extraNotContains: []any{42},
		//	options:          []SchemaValidationOption{VisitAsResponse()},
		//},
		{
			name: "required properties",
			schema: &Schema{
				Properties: Schemas{
					"bar": NewStringSchema().NewRef(),
				},
				Required: []string{"bar"},
			},
			value:            map[string]any{"foo": 42},
			extraNotContains: []any{42},
		},
		{
			name: "one of (matches more then one)",
			schema: NewOneOfSchema(
				&Schema{MultipleOf: Ptr(6.0)},
				&Schema{MultipleOf: Ptr(7.0)},
			),
			value: 42,
		},
		{
			name: "one of (no matches)",
			schema: NewOneOfSchema(
				&Schema{MultipleOf: Ptr(5.0)},
				&Schema{MultipleOf: Ptr(10.0)},
			),
			value: 42,
		},
		{
			name: "any of",
			schema: NewAnyOfSchema(
				&Schema{MultipleOf: Ptr(5.0)},
				&Schema{MultipleOf: Ptr(10.0)},
			),
			value: 42,
		},
		{
			name: "all of (match some)",
			schema: NewAllOfSchema(
				&Schema{MultipleOf: Ptr(6.0)},
				&Schema{MultipleOf: Ptr(5.0)},
			),
			value: 42,
		},
		{
			name: "all of (no match)",
			schema: NewAllOfSchema(
				&Schema{MultipleOf: Ptr(10.0)},
				&Schema{MultipleOf: Ptr(5.0)},
			),
			value: 42,
		},
		{
			name:   "uuid format",
			schema: NewUUIDSchema(),
			value:  "foo",
		},
		{
			name:   "date time format",
			schema: NewDateTimeSchema(),
			value:  "foo",
		},
		{
			name:   "date format",
			schema: NewSchema().WithFormat("date"),
			value:  "foo",
		},
		{
			name:   "ipv4 format",
			schema: NewSchema().WithFormat("ipv4"),
			value:  "foo",
		},
		{
			name:   "ipv6 format",
			schema: NewSchema().WithFormat("ipv6"),
			value:  "foo",
		},
		{
			name:   "email format",
			schema: NewSchema().WithFormat("email"),
			value:  "foo",
		},
		{
			name:   "byte format",
			schema: NewBytesSchema(),
			value:  "foo!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.schema.VisitJSON(tc.value, tc.options...)
			var schemaError = &SchemaError{}
			require.Error(t, err)
			require.ErrorAs(t, err, &schemaError)
			require.NotZero(t, schemaError.Reason)
			require.NotContains(t, schemaError.Reason, fmt.Sprint(tc.value))
			for _, extra := range tc.extraNotContains {
				require.NotContains(t, schemaError.Reason, fmt.Sprint(extra))
			}
		})
	}
}
