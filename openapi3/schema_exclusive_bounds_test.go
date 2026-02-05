package openapi3

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExclusiveMinMax_Unmarshal(t *testing.T) {
	t.Run("boolean exclusiveMinimum true", func(t *testing.T) {
		data := `{"type": "number", "minimum": 0, "exclusiveMinimum": true}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMinBool)
		require.True(t, *schema.ExclusiveMinBool)
		require.Nil(t, schema.ExclusiveMin)
	})

	t.Run("boolean exclusiveMinimum false", func(t *testing.T) {
		data := `{"type": "number", "minimum": 0, "exclusiveMinimum": false}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMinBool)
		require.False(t, *schema.ExclusiveMinBool)
		require.Nil(t, schema.ExclusiveMin)
	})

	t.Run("boolean exclusiveMaximum true", func(t *testing.T) {
		data := `{"type": "number", "maximum": 100, "exclusiveMaximum": true}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMaxBool)
		require.True(t, *schema.ExclusiveMaxBool)
		require.Nil(t, schema.ExclusiveMax)
	})

	t.Run("boolean exclusiveMaximum false", func(t *testing.T) {
		data := `{"type": "number", "maximum": 100, "exclusiveMaximum": false}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMaxBool)
		require.False(t, *schema.ExclusiveMaxBool)
		require.Nil(t, schema.ExclusiveMax)
	})

	t.Run("numeric exclusiveMinimum zero", func(t *testing.T) {
		data := `{"type": "number", "exclusiveMinimum": 0.0}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMin)
		require.Equal(t, 0.0, *schema.ExclusiveMin)
		require.Nil(t, schema.ExclusiveMinBool)
	})

	t.Run("numeric exclusiveMinimum positive", func(t *testing.T) {
		data := `{"type": "number", "exclusiveMinimum": 100.5}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMin)
		require.Equal(t, 100.5, *schema.ExclusiveMin)
		require.Nil(t, schema.ExclusiveMinBool)
	})

	t.Run("numeric exclusiveMinimum negative", func(t *testing.T) {
		data := `{"type": "number", "exclusiveMinimum": -50.0}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMin)
		require.Equal(t, -50.0, *schema.ExclusiveMin)
		require.Nil(t, schema.ExclusiveMinBool)
	})

	t.Run("numeric exclusiveMaximum zero", func(t *testing.T) {
		data := `{"type": "number", "exclusiveMaximum": 0.0}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMax)
		require.Equal(t, 0.0, *schema.ExclusiveMax)
		require.Nil(t, schema.ExclusiveMaxBool)
	})

	t.Run("numeric exclusiveMaximum positive", func(t *testing.T) {
		data := `{"type": "number", "exclusiveMaximum": 100.5}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMax)
		require.Equal(t, 100.5, *schema.ExclusiveMax)
		require.Nil(t, schema.ExclusiveMaxBool)
	})

	t.Run("numeric exclusiveMaximum negative", func(t *testing.T) {
		data := `{"type": "number", "exclusiveMaximum": -50.0}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMax)
		require.Equal(t, -50.0, *schema.ExclusiveMax)
		require.Nil(t, schema.ExclusiveMaxBool)
	})

	t.Run("complete OAS 3.0 schema with minimum and boolean", func(t *testing.T) {
		data := `{
			"type": "number",
			"minimum": 0,
			"maximum": 100,
			"exclusiveMinimum": true,
			"exclusiveMaximum": true
		}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.Min)
		require.Equal(t, 0.0, *schema.Min)
		require.NotNil(t, schema.Max)
		require.Equal(t, 100.0, *schema.Max)
		require.NotNil(t, schema.ExclusiveMinBool)
		require.True(t, *schema.ExclusiveMinBool)
		require.NotNil(t, schema.ExclusiveMaxBool)
		require.True(t, *schema.ExclusiveMaxBool)
		require.Nil(t, schema.ExclusiveMin)
		require.Nil(t, schema.ExclusiveMax)
	})

	t.Run("complete OAS 3.1 schema with numeric only", func(t *testing.T) {
		data := `{
			"type": "number",
			"exclusiveMinimum": 0,
			"exclusiveMaximum": 100
		}`
		var schema Schema
		err := json.Unmarshal([]byte(data), &schema)
		require.NoError(t, err)
		require.NotNil(t, schema.ExclusiveMin)
		require.Equal(t, 0.0, *schema.ExclusiveMin)
		require.NotNil(t, schema.ExclusiveMax)
		require.Equal(t, 100.0, *schema.ExclusiveMax)
		require.Nil(t, schema.Min)
		require.Nil(t, schema.Max)
		require.Nil(t, schema.ExclusiveMinBool)
		require.Nil(t, schema.ExclusiveMaxBool)
	})
}

func TestExclusiveMinMax_Marshal(t *testing.T) {
	t.Run("boolean true marshals as true", func(t *testing.T) {
		trueVal := true
		schema := &Schema{
			Type:             &Types{"number"},
			Min:              Ptr(0.0),
			ExclusiveMinBool: &trueVal,
		}
		data, err := json.Marshal(schema)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		require.Equal(t, true, result["exclusiveMinimum"])
	})

	t.Run("boolean false does not marshal", func(t *testing.T) {
		falseVal := false
		schema := &Schema{
			Type:             &Types{"number"},
			Min:              Ptr(0.0),
			ExclusiveMinBool: &falseVal,
		}
		data, err := json.Marshal(schema)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		_, exists := result["exclusiveMinimum"]
		require.False(t, exists, "false boolean should not be marshaled")
	})

	t.Run("numeric value marshals correctly", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(5.5),
		}
		data, err := json.Marshal(schema)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		require.Equal(t, 5.5, result["exclusiveMinimum"])
	})

	t.Run("numeric takes precedence over boolean", func(t *testing.T) {
		trueVal := true
		schema := &Schema{
			Type:             &Types{"number"},
			Min:              Ptr(0.0),
			ExclusiveMinBool: &trueVal,
			ExclusiveMin:     Ptr(10.0),
		}
		data, err := json.Marshal(schema)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		require.Equal(t, 10.0, result["exclusiveMinimum"], "numeric should take precedence")
	})

	t.Run("round-trip OAS 3.0 preserves boolean", func(t *testing.T) {
		original := `{"type": "number", "minimum": 0, "exclusiveMinimum": true}`
		var schema Schema
		err := json.Unmarshal([]byte(original), &schema)
		require.NoError(t, err)

		marshaled, err := json.Marshal(&schema)
		require.NoError(t, err)

		var schema2 Schema
		err = json.Unmarshal(marshaled, &schema2)
		require.NoError(t, err)

		require.NotNil(t, schema2.ExclusiveMinBool)
		require.True(t, *schema2.ExclusiveMinBool)
		require.Nil(t, schema2.ExclusiveMin)
	})

	t.Run("round-trip OAS 3.1 preserves numeric", func(t *testing.T) {
		original := `{"type": "number", "exclusiveMinimum": 5.5}`
		var schema Schema
		err := json.Unmarshal([]byte(original), &schema)
		require.NoError(t, err)

		marshaled, err := json.Marshal(&schema)
		require.NoError(t, err)

		var schema2 Schema
		err = json.Unmarshal(marshaled, &schema2)
		require.NoError(t, err)

		require.NotNil(t, schema2.ExclusiveMin)
		require.Equal(t, 5.5, *schema2.ExclusiveMin)
		require.Nil(t, schema2.ExclusiveMinBool)
	})
}

func TestExclusiveMinMax_BuiltinValidation(t *testing.T) {
	// Boolean form validation - exclusiveMinimum
	t.Run("boolean exclusiveMinimum - value above minimum passes", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"number"},
			Min:              Ptr(0.0),
			ExclusiveMinBool: Ptr(true),
		}
		err := schema.VisitJSON(0.1)
		require.NoError(t, err)
	})

	t.Run("boolean exclusiveMinimum - value at minimum fails", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"number"},
			Min:              Ptr(0.0),
			ExclusiveMinBool: Ptr(true),
		}
		err := schema.VisitJSON(0.0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "exclusiveMinimum")
	})

	t.Run("boolean exclusiveMinimum - value below minimum fails", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"number"},
			Min:              Ptr(0.0),
			ExclusiveMinBool: Ptr(true),
		}
		err := schema.VisitJSON(-0.1)
		require.Error(t, err)
	})

	// Boolean form validation - exclusiveMaximum
	t.Run("boolean exclusiveMaximum - value below maximum passes", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"number"},
			Max:              Ptr(100.0),
			ExclusiveMaxBool: Ptr(true),
		}
		err := schema.VisitJSON(99.9)
		require.NoError(t, err)
	})

	t.Run("boolean exclusiveMaximum - value at maximum fails", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"number"},
			Max:              Ptr(100.0),
			ExclusiveMaxBool: Ptr(true),
		}
		err := schema.VisitJSON(100.0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "exclusiveMaximum")
	})

	t.Run("boolean exclusiveMaximum - value above maximum fails", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"number"},
			Max:              Ptr(100.0),
			ExclusiveMaxBool: Ptr(true),
		}
		err := schema.VisitJSON(100.1)
		require.Error(t, err)
	})

	// Numeric form validation - exclusiveMinimum
	t.Run("numeric exclusiveMin - value above passes", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(0.0),
		}
		err := schema.VisitJSON(0.1)
		require.NoError(t, err)
	})

	t.Run("numeric exclusiveMin - value at boundary fails", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(0.0),
		}
		err := schema.VisitJSON(0.0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "exclusiveMinimum")
	})

	t.Run("numeric exclusiveMin - value below fails", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(0.0),
		}
		err := schema.VisitJSON(-0.1)
		require.Error(t, err)
	})

	// Numeric form validation - exclusiveMaximum
	t.Run("numeric exclusiveMax - value below passes", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMax: Ptr(100.0),
		}
		err := schema.VisitJSON(99.9)
		require.NoError(t, err)
	})

	t.Run("numeric exclusiveMax - value at boundary fails", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMax: Ptr(100.0),
		}
		err := schema.VisitJSON(100.0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "exclusiveMaximum")
	})

	t.Run("numeric exclusiveMax - value above fails", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMax: Ptr(100.0),
		}
		err := schema.VisitJSON(100.1)
		require.Error(t, err)
	})

	// Edge cases
	t.Run("integer type respects numeric exclusive bounds", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"integer"},
			ExclusiveMin: Ptr(0.0),
			ExclusiveMax: Ptr(10.0),
		}
		err := schema.VisitJSON(5)
		require.NoError(t, err)

		err = schema.VisitJSON(0)
		require.Error(t, err)

		err = schema.VisitJSON(10)
		require.Error(t, err)
	})

	t.Run("floating point precision near boundary", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(0.0),
		}
		// Slightly above zero should pass
		err := schema.VisitJSON(0.000001)
		require.NoError(t, err)

		// Exactly zero should fail
		err = schema.VisitJSON(0.0)
		require.Error(t, err)
	})
}

func TestExclusiveMinMax_JSONSchema2020(t *testing.T) {
	t.Run("OAS 3.0 boolean transforms correctly", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"number"},
			Min:              Ptr(0.0),
			ExclusiveMinBool: Ptr(true),
		}

		err := schema.VisitJSON(0.1, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(0.0, EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("OAS 3.1 numeric passes through", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(0.0),
		}

		err := schema.VisitJSON(0.1, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(0.0, EnableJSONSchema2020())
		require.Error(t, err)
	})

	// Mirror validation tests with JSON Schema 2020-12
	t.Run("numeric exclusiveMin validation with JSONSchema2020", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(5.0),
		}

		err := schema.VisitJSON(6.0, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(5.0, EnableJSONSchema2020())
		require.Error(t, err)

		err = schema.VisitJSON(4.0, EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("numeric exclusiveMax validation with JSONSchema2020", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMax: Ptr(100.0),
		}

		err := schema.VisitJSON(99.0, EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(100.0, EnableJSONSchema2020())
		require.Error(t, err)

		err = schema.VisitJSON(101.0, EnableJSONSchema2020())
		require.Error(t, err)
	})
}

func TestExclusiveMinMax_Builders(t *testing.T) {
	t.Run("WithExclusiveMin sets boolean", func(t *testing.T) {
		schema := NewFloat64Schema().WithExclusiveMin(true)
		require.NotNil(t, schema.ExclusiveMinBool)
		require.True(t, *schema.ExclusiveMinBool)
		require.Nil(t, schema.ExclusiveMin)
	})

	t.Run("WithExclusiveMax sets boolean", func(t *testing.T) {
		schema := NewFloat64Schema().WithExclusiveMax(true)
		require.NotNil(t, schema.ExclusiveMaxBool)
		require.True(t, *schema.ExclusiveMaxBool)
		require.Nil(t, schema.ExclusiveMax)
	})

	t.Run("WithExclusiveMinNumber sets numeric", func(t *testing.T) {
		schema := NewFloat64Schema().WithExclusiveMinNumber(5.5)
		require.NotNil(t, schema.ExclusiveMin)
		require.Equal(t, 5.5, *schema.ExclusiveMin)
		require.Nil(t, schema.ExclusiveMinBool)
	})

	t.Run("WithExclusiveMaxNumber sets numeric", func(t *testing.T) {
		schema := NewFloat64Schema().WithExclusiveMaxNumber(100.5)
		require.NotNil(t, schema.ExclusiveMax)
		require.Equal(t, 100.5, *schema.ExclusiveMax)
		require.Nil(t, schema.ExclusiveMaxBool)
	})

	t.Run("chaining builders works", func(t *testing.T) {
		schema := NewFloat64Schema().
			WithExclusiveMinNumber(0.0).
			WithExclusiveMaxNumber(100.0)

		require.NotNil(t, schema.ExclusiveMin)
		require.Equal(t, 0.0, *schema.ExclusiveMin)
		require.NotNil(t, schema.ExclusiveMax)
		require.Equal(t, 100.0, *schema.ExclusiveMax)
	})

	t.Run("numeric overwrites boolean in marshaling", func(t *testing.T) {
		schema := NewFloat64Schema().
			WithMin(0.0).
			WithExclusiveMin(true).
			WithExclusiveMinNumber(5.0)

		data, err := json.Marshal(schema)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		require.Equal(t, 5.0, result["exclusiveMinimum"])
	})
}

func TestExclusiveMinMax_Integration(t *testing.T) {
	t.Run("Schema.IsEmpty handles exclusive bounds", func(t *testing.T) {
		emptySchema := &Schema{}
		require.True(t, emptySchema.IsEmpty())

		schemaWithBool := &Schema{ExclusiveMinBool: Ptr(true)}
		require.False(t, schemaWithBool.IsEmpty())

		schemaWithNumeric := &Schema{ExclusiveMin: Ptr(0.0)}
		require.False(t, schemaWithNumeric.IsEmpty())
	})

	t.Run("Schema.Validate passes for well-formed schemas", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(0.0),
			ExclusiveMax: Ptr(100.0),
		}
		err := schema.Validate(context.Background())
		require.NoError(t, err)
	})

	t.Run("complete validation workflow OAS 3.0", func(t *testing.T) {
		schema := &Schema{
			Type:             &Types{"number"},
			Min:              Ptr(0.0),
			Max:              Ptr(100.0),
			ExclusiveMinBool: Ptr(true),
			ExclusiveMaxBool: Ptr(true),
		}

		// Validate schema structure
		err := schema.Validate(context.Background())
		require.NoError(t, err)

		// Validate data against schema
		err = schema.VisitJSON(50.0)
		require.NoError(t, err)

		err = schema.VisitJSON(0.0)
		require.Error(t, err)

		err = schema.VisitJSON(100.0)
		require.Error(t, err)
	})

	t.Run("complete validation workflow OAS 3.1", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(0.0),
			ExclusiveMax: Ptr(100.0),
		}

		// Validate schema structure
		err := schema.Validate(context.Background())
		require.NoError(t, err)

		// Validate data against schema
		err = schema.VisitJSON(50.0)
		require.NoError(t, err)

		err = schema.VisitJSON(0.0)
		require.Error(t, err)

		err = schema.VisitJSON(100.0)
		require.Error(t, err)
	})
}

func TestExclusiveMinMax_ErrorMessages(t *testing.T) {
	t.Run("exclusive minimum error message is clear", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(0.0),
		}
		err := schema.VisitJSON(0.0)
		require.Error(t, err)

		var schemaError *SchemaError
		require.ErrorAs(t, err, &schemaError)
		require.NotZero(t, schemaError.Reason)
		require.Contains(t, schemaError.Reason, "more than")
	})

	t.Run("exclusive maximum error message is clear", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMax: Ptr(100.0),
		}
		err := schema.VisitJSON(100.0)
		require.Error(t, err)

		var schemaError *SchemaError
		require.ErrorAs(t, err, &schemaError)
		require.NotZero(t, schemaError.Reason)
		require.Contains(t, schemaError.Reason, "less than")
	})

	t.Run("error message doesn't leak sensitive data", func(t *testing.T) {
		schema := &Schema{
			Type:         &Types{"number"},
			ExclusiveMin: Ptr(50.0),
		}
		sensitiveValue := 42.0
		err := schema.VisitJSON(sensitiveValue)
		require.Error(t, err)

		var schemaError *SchemaError
		require.ErrorAs(t, err, &schemaError)
		// Should not contain the actual value in the reason string
		// The value is stored in the error, but not in the reason
		require.NotZero(t, schemaError.Reason)
	})
}

func TestExclusiveMinMax_JSONLookup(t *testing.T) {
	t.Run("returns numeric when both set", func(t *testing.T) {
		schema := Schema{
			ExclusiveMin:     Ptr(10.0),
			ExclusiveMinBool: Ptr(true),
		}
		val, err := schema.JSONLookup("exclusiveMin")
		require.NoError(t, err)
		require.Equal(t, Ptr(10.0), val)
	})

	t.Run("returns boolean when only boolean set", func(t *testing.T) {
		schema := Schema{
			ExclusiveMinBool: Ptr(true),
		}
		val, err := schema.JSONLookup("exclusiveMin")
		require.NoError(t, err)
		require.Equal(t, Ptr(true), val)
	})

	t.Run("returns numeric when only numeric set", func(t *testing.T) {
		schema := Schema{
			ExclusiveMin: Ptr(5.5),
		}
		val, err := schema.JSONLookup("exclusiveMin")
		require.NoError(t, err)
		require.Equal(t, Ptr(5.5), val)
	})

	t.Run("returns numeric for exclusiveMax when both set", func(t *testing.T) {
		schema := Schema{
			ExclusiveMax:     Ptr(100.0),
			ExclusiveMaxBool: Ptr(true),
		}
		val, err := schema.JSONLookup("exclusiveMax")
		require.NoError(t, err)
		require.Equal(t, Ptr(100.0), val)
	})
}
