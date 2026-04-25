package openapi3

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue430(t *testing.T) {
	schema := NewOneOfSchema(
		NewStringSchema().WithFormat("ipv4"),
		NewStringSchema().WithFormat("ipv6"),
	)

	delete(SchemaStringFormats, "ipv4")
	delete(SchemaStringFormats, "ipv6")

	err := schema.Validate(t.Context())
	require.NoError(t, err)

	data := map[string]bool{
		"127.0.1.1": true,

		// https://stackoverflow.com/a/48519490/1418165

		// v4
		"192.168.0.1": true,
		// "192.168.0.1:80" doesn't parse per net.ParseIP()

		// v6
		"::FFFF:C0A8:1":                        false,
		"::FFFF:C0A8:0001":                     false,
		"0000:0000:0000:0000:0000:FFFF:C0A8:1": false,
		// "::FFFF:C0A8:1%1" doesn't parse per net.ParseIP()
		"::FFFF:192.168.0.1": false,
		// "[::FFFF:C0A8:1]:80" doesn't parse per net.ParseIP()
		// "[::FFFF:C0A8:1%1]:80" doesn't parse per net.ParseIP()
		"2001:db8::": false,
	}

	for datum := range data {
		err = schema.VisitJSON(datum)
		require.Error(t, err, ErrOneOfConflict.Error())
	}

	DefineIPv4Format()
	DefineIPv6Format()

	ipv4Validator := NewIPValidator(true)
	ipv6Validator := NewIPValidator(false)

	for datum, isV4 := range data {
		err = schema.VisitJSON(datum)
		require.NoError(t, err)
		if isV4 {
			assert.Nil(t, ipv4Validator.Validate(datum), "%q should be IPv4", datum)
			assert.NotNil(t, ipv6Validator.Validate(datum), "%q should not be IPv6", datum)
		} else {
			assert.NotNil(t, ipv4Validator.Validate(datum), "%q should not be IPv4", datum)
			assert.Nil(t, ipv6Validator.Validate(datum), "%q should be IPv6", datum)
		}
	}
}

func TestFormatCallback_WrapError(t *testing.T) {
	var errSomething = errors.New("something error")

	DefineStringFormatValidator("foobar", NewCallbackValidator(func(value string) error {
		return errSomething
	}))

	s := &Schema{Format: "foobar"}
	err := s.VisitJSONString("blablabla")

	assert.ErrorIs(t, err, errSomething)

	delete(SchemaStringFormats, "foobar")
}

func TestReversePathInMessageSchemaError(t *testing.T) {
	DefineIPv4Format()

	SchemaErrorDetailsDisabled = true

	const spc = `
components:
  schemas:
    Something:
      type: object
      properties:
        ip:
          type: string
          format: ipv4
`
	l := NewLoader()

	doc, err := l.LoadFromData([]byte(spc))
	require.NoError(t, err)

	err = doc.Components.Schemas["Something"].Value.VisitJSON(map[string]any{
		`ip`: `123.0.0.11111`,
	})

	require.EqualError(t, err, `Error at "/ip": string doesn't match the format "ipv4": Not an IP address`)

	delete(SchemaStringFormats, "ipv4")
	SchemaErrorDetailsDisabled = false
}

func TestUuidFormat(t *testing.T) {

	type testCase struct {
		name    string
		value   string
		wantErr bool
	}

	DefineStringFormatValidator("uuid", NewRegexpFormatValidator(FormatOfStringForUUIDOfRFC4122))
	testCases := []testCase{
		{
			name:    "invalid",
			value:   "foo",
			wantErr: true,
		},
		{
			name:    "uuid v1",
			value:   "77e66540-ca29-11ed-afa1-0242ac120002",
			wantErr: false,
		},
		{
			name:    "uuid v4",
			value:   "00f4d301-b9f4-4366-8907-2b5a03430aa1",
			wantErr: false,
		},
		{
			name:    "uuid nil",
			value:   "00000000-0000-0000-0000-000000000000",
			wantErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewUUIDSchema().VisitJSON(tc.value)
			var schemaError = &SchemaError{}
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &schemaError)

				require.NotZero(t, schemaError.Reason)
				require.NotContains(t, schemaError.Reason, fmt.Sprint(tc.value))
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestNumberFormats(t *testing.T) {
	type testCase struct {
		name    string
		typ     string
		format  string
		value   any
		wantErr bool
	}
	DefineNumberFormatValidator("lessThan10", NewCallbackValidator(func(value float64) error {
		if value >= 10 {
			return errors.New("not less than 10")
		}
		return nil
	}))
	DefineIntegerFormatValidator("odd", NewCallbackValidator(func(value int64) error {
		if value%2 == 0 {
			return errors.New("not odd")
		}
		return nil
	}))
	testCases := []testCase{
		{
			name:    "invalid number",
			value:   "test",
			typ:     "number",
			format:  "",
			wantErr: true,
		},
		{
			name:    "zero float64",
			value:   0.0,
			typ:     "number",
			format:  "lessThan10",
			wantErr: false,
		},
		{
			name:    "11",
			value:   11.0,
			typ:     "number",
			format:  "lessThan10",
			wantErr: true,
		},
		{
			name:    "odd 11",
			value:   11.0,
			typ:     "integer",
			format:  "odd",
			wantErr: false,
		},
		{
			name:    "even 12",
			value:   12.0,
			typ:     "integer",
			format:  "odd",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema := &Schema{
				Type:   &Types{tc.typ},
				Format: tc.format,
			}
			err := schema.VisitJSON(tc.value)
			var schemaError = &SchemaError{}
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &schemaError)

				require.NotZero(t, schemaError.Reason)
				require.NotContains(t, schemaError.Reason, fmt.Sprint(tc.value))
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestPerValidationFormatValidators(t *testing.T) {
	// This test demonstrates per-validation format validators that don't pollute global state
	// and allow different validators for the same format name across different specs

	// Create two different validators for the same format name
	validatorA := NewCallbackValidator(func(value string) error {
		if value != "spec-a-value" {
			return errors.New("not a valid spec-a value")
		}
		return nil
	})

	validatorB := NewCallbackValidator(func(value string) error {
		if value != "spec-b-value" {
			return errors.New("not a valid spec-b value")
		}
		return nil
	})

	schema := &Schema{
		Type:   &Types{"string"},
		Format: "custom-id",
	}

	// Test with validator A
	err := schema.VisitJSON("spec-a-value", WithStringFormatValidator("custom-id", validatorA))
	require.NoError(t, err)

	err = schema.VisitJSON("spec-b-value", WithStringFormatValidator("custom-id", validatorA))
	require.Error(t, err)
	require.ErrorContains(t, err, "not a valid spec-a value")

	// Test with validator B - completely independent
	err = schema.VisitJSON("spec-b-value", WithStringFormatValidator("custom-id", validatorB))
	require.NoError(t, err)

	err = schema.VisitJSON("spec-a-value", WithStringFormatValidator("custom-id", validatorB))
	require.Error(t, err)
	require.ErrorContains(t, err, "not a valid spec-b value")

	// Verify global validators are not affected
	_, globalExists := SchemaStringFormats["custom-id"]
	require.False(t, globalExists, "global SchemaStringFormats should not be polluted")
}

func TestPerValidationFormatValidators_AllTypes(t *testing.T) {
	// Test string format validator
	stringValidator := NewCallbackValidator(func(value string) error {
		if len(value) < 5 {
			return errors.New("string too short")
		}
		return nil
	})

	stringSchema := &Schema{
		Type:   &Types{"string"},
		Format: "min5",
	}

	err := stringSchema.VisitJSON("hello", WithStringFormatValidator("min5", stringValidator))
	require.NoError(t, err)

	err = stringSchema.VisitJSON("hi", WithStringFormatValidator("min5", stringValidator))
	require.Error(t, err)

	// Test number format validator
	numberValidator := NewCallbackValidator(func(value float64) error {
		if value < 0 {
			return errors.New("must be non-negative")
		}
		return nil
	})

	numberSchema := &Schema{
		Type:   &Types{"number"},
		Format: "non-negative",
	}

	err = numberSchema.VisitJSON(5.5, WithNumberFormatValidator("non-negative", numberValidator))
	require.NoError(t, err)

	err = numberSchema.VisitJSON(-1.0, WithNumberFormatValidator("non-negative", numberValidator))
	require.Error(t, err)

	// Test integer format validator
	integerValidator := NewCallbackValidator(func(value int64) error {
		if value%2 != 0 {
			return errors.New("must be even")
		}
		return nil
	})

	integerSchema := &Schema{
		Type:   &Types{"integer"},
		Format: "even",
	}

	err = integerSchema.VisitJSON(float64(4), WithIntegerFormatValidator("even", integerValidator))
	require.NoError(t, err)

	err = integerSchema.VisitJSON(float64(5), WithIntegerFormatValidator("even", integerValidator))
	require.Error(t, err)
}

func TestPerValidationFormatValidators_FallbackToGlobal(t *testing.T) {
	// Register a global validator
	DefineStringFormatValidator("global-format", NewCallbackValidator(func(value string) error {
		if value != "global" {
			return errors.New("not global value")
		}
		return nil
	}))
	defer delete(SchemaStringFormats, "global-format")

	schema := &Schema{
		Type:   &Types{"string"},
		Format: "global-format",
	}

	// Should use global validator when no per-validation validator is provided
	err := schema.VisitJSON("global")
	require.NoError(t, err)

	err = schema.VisitJSON("other")
	require.Error(t, err)

	// Per-validation validator should override global
	localValidator := NewCallbackValidator(func(value string) error {
		if value != "local" {
			return errors.New("not local value")
		}
		return nil
	})

	err = schema.VisitJSON("local", WithStringFormatValidator("global-format", localValidator))
	require.NoError(t, err)

	err = schema.VisitJSON("global", WithStringFormatValidator("global-format", localValidator))
	require.Error(t, err)
}

func TestDocumentScopedFormatValidators(t *testing.T) {
	// This test demonstrates document-scoped format validators
	// Different OpenAPI specs can have different validators for the same format name

	// Create Spec A with its own validation rules
	specA := &T{
		OpenAPI: "3.0.0",
		Info:    &Info{Title: "Spec A", Version: "1.0.0"},
		Paths:   NewPaths(),
	}

	validatorA := NewCallbackValidator(func(value string) error {
		if len(value) < 2 || value[:2] != "A-" {
			return errors.New("must start with 'A-'")
		}
		return nil
	})
	specA.SetStringFormatValidator("custom-id", validatorA)

	// Create Spec B with different validation rules for the same format
	specB := &T{
		OpenAPI: "3.0.0",
		Info:    &Info{Title: "Spec B", Version: "1.0.0"},
		Paths:   NewPaths(),
	}

	validatorB := NewCallbackValidator(func(value string) error {
		if len(value) < 2 || value[:2] != "B-" {
			return errors.New("must start with 'B-'")
		}
		return nil
	})
	specB.SetStringFormatValidator("custom-id", validatorB)

	// Create a schema that uses the custom-id format
	schema := &Schema{
		Type:   &Types{"string"},
		Format: "custom-id",
	}

	// Validate against Spec A - using GetSchemaValidationOptions
	err := schema.VisitJSON("A-123", specA.GetSchemaValidationOptions()...)
	require.NoError(t, err)

	err = schema.VisitJSON("B-456", specA.GetSchemaValidationOptions()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "must start with 'A-'")

	// Validate against Spec B - completely independent
	err = schema.VisitJSON("B-456", specB.GetSchemaValidationOptions()...)
	require.NoError(t, err)

	err = schema.VisitJSON("A-123", specB.GetSchemaValidationOptions()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "must start with 'B-'")

	// Or use the convenience method ValidateSchemaJSON
	err = specA.ValidateSchemaJSON(schema, "A-789")
	require.NoError(t, err)

	err = specB.ValidateSchemaJSON(schema, "B-789")
	require.NoError(t, err)

	// Verify global validators are not affected
	_, globalExists := SchemaStringFormats["custom-id"]
	require.False(t, globalExists, "global SchemaStringFormats should not be polluted")
}

func TestDocumentScopedFormatValidators_AllTypes(t *testing.T) {
	doc := &T{
		OpenAPI: "3.0.0",
		Info:    &Info{Title: "Test", Version: "1.0.0"},
		Paths:   NewPaths(),
	}

	// Set string validator
	doc.SetStringFormatValidator("min5", NewCallbackValidator(func(value string) error {
		if len(value) < 5 {
			return errors.New("string too short")
		}
		return nil
	}))

	// Set number validator
	doc.SetNumberFormatValidator("non-negative", NewCallbackValidator(func(value float64) error {
		if value < 0 {
			return errors.New("must be non-negative")
		}
		return nil
	}))

	// Set integer validator
	doc.SetIntegerFormatValidator("even", NewCallbackValidator(func(value int64) error {
		if value%2 != 0 {
			return errors.New("must be even")
		}
		return nil
	}))

	// Test string
	stringSchema := &Schema{Type: &Types{"string"}, Format: "min5"}
	err := doc.ValidateSchemaJSON(stringSchema, "hello")
	require.NoError(t, err)
	err = doc.ValidateSchemaJSON(stringSchema, "hi")
	require.Error(t, err)

	// Test number
	numberSchema := &Schema{Type: &Types{"number"}, Format: "non-negative"}
	err = doc.ValidateSchemaJSON(numberSchema, 5.5)
	require.NoError(t, err)
	err = doc.ValidateSchemaJSON(numberSchema, -1.0)
	require.Error(t, err)

	// Test integer
	integerSchema := &Schema{Type: &Types{"integer"}, Format: "even"}
	err = doc.ValidateSchemaJSON(integerSchema, float64(4))
	require.NoError(t, err)
	err = doc.ValidateSchemaJSON(integerSchema, float64(5))
	require.Error(t, err)
}

func TestDocumentScopedFormatValidators_BatchSet(t *testing.T) {
	doc := &T{
		OpenAPI: "3.0.0",
		Info:    &Info{Title: "Test", Version: "1.0.0"},
		Paths:   NewPaths(),
	}

	// Set multiple validators at once
	validators := map[string]StringFormatValidator{
		"format1": NewCallbackValidator(func(value string) error {
			if value != "valid1" {
				return errors.New("not valid1")
			}
			return nil
		}),
		"format2": NewCallbackValidator(func(value string) error {
			if value != "valid2" {
				return errors.New("not valid2")
			}
			return nil
		}),
	}
	doc.SetStringFormatValidators(validators)

	schema1 := &Schema{Type: &Types{"string"}, Format: "format1"}
	schema2 := &Schema{Type: &Types{"string"}, Format: "format2"}

	err := doc.ValidateSchemaJSON(schema1, "valid1")
	require.NoError(t, err)

	err = doc.ValidateSchemaJSON(schema2, "valid2")
	require.NoError(t, err)

	err = doc.ValidateSchemaJSON(schema1, "valid2")
	require.Error(t, err)
}
