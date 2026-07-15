package openapi3filter

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// Tests for (*urlValuesDecoder).DecodeArray across form/space/pipe styles
// and explode settings, including single-value and delimited multi-value inputs.

func decodeArrayStringSchema() *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type:  &openapi3.Types{"array"},
		Items: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
	}}
}

func decodeArrayIntSchema() *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type:  &openapi3.Types{"array"},
		Items: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
	}}
}

func TestDecodeArray_FormNoExplode(t *testing.T) {
	sm := &openapi3.SerializationMethod{Style: "form", Explode: false}
	schema := decodeArrayStringSchema()

	t.Run("single value no delimiter", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo"}, got)
	})

	t.Run("multi value comma delimited", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo,bar,baz"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo", "bar", "baz"}, got)
	})

	t.Run("trailing delimiter", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo,"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo", ""}, got)
	})

	t.Run("leading delimiter", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {",foo"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"", "foo"}, got)
	})

	t.Run("empty middle segment", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"a,,b"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"a", "", "b"}, got)
	})

	t.Run("empty string only", func(t *testing.T) {
		// parseArray treats a single empty string as "no value".
		dec := &urlValuesDecoder{values: url.Values{"param": {""}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Nil(t, got)
	})

	t.Run("missing param", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.False(t, found)
		require.Nil(t, got)
	})

	t.Run("pipe chars are not delimiters in form style", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo|bar"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo|bar"}, got)
	})
}

func TestDecodeArray_SpaceAndPipeNoExplode(t *testing.T) {
	schema := decodeArrayStringSchema()

	t.Run("spaceDelimited single", func(t *testing.T) {
		sm := &openapi3.SerializationMethod{Style: "spaceDelimited", Explode: false}
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo"}, got)
	})

	t.Run("spaceDelimited multi", func(t *testing.T) {
		sm := &openapi3.SerializationMethod{Style: "spaceDelimited", Explode: false}
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo bar baz"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo", "bar", "baz"}, got)
	})

	t.Run("pipeDelimited single", func(t *testing.T) {
		sm := &openapi3.SerializationMethod{Style: "pipeDelimited", Explode: false}
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo"}, got)
	})

	t.Run("pipeDelimited multi", func(t *testing.T) {
		sm := &openapi3.SerializationMethod{Style: "pipeDelimited", Explode: false}
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo|bar|baz"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo", "bar", "baz"}, got)
	})
}

func TestDecodeArray_FormExplode(t *testing.T) {
	sm := &openapi3.SerializationMethod{Style: "form", Explode: true}
	schema := decodeArrayStringSchema()

	t.Run("single repeated key", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo"}, got)
	})

	t.Run("multiple repeated keys", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo", "bar", "baz"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo", "bar", "baz"}, got)
	})

	t.Run("commas are part of the value when explode", func(t *testing.T) {
		// explode skips Split; commas must not become item separators.
		dec := &urlValuesDecoder{values: url.Values{"param": {"foo,bar"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{"foo,bar"}, got)
	})
}

func TestDecodeArray_IntegerItems(t *testing.T) {
	sm := &openapi3.SerializationMethod{Style: "form", Explode: false}
	schema := decodeArrayIntSchema()

	t.Run("single", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"42"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{int64(42)}, got)
	})

	t.Run("multi", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"1,2,3"}}}
		got, found, err := dec.DecodeArray("param", sm, schema)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, []any{int64(1), int64(2), int64(3)}, got)
	})

	t.Run("invalid item", func(t *testing.T) {
		dec := &urlValuesDecoder{values: url.Values{"param": {"1,x"}}}
		_, found, err := dec.DecodeArray("param", sm, schema)
		require.True(t, found)
		require.Error(t, err)
		var pe *ParseError
		require.ErrorAs(t, err, &pe)
		require.Equal(t, []any{1}, pe.Path())
	})
}

func TestDecodeArray_DeepObjectRejected(t *testing.T) {
	sm := &openapi3.SerializationMethod{Style: "deepObject", Explode: true}
	dec := &urlValuesDecoder{values: url.Values{"param": {"x"}}}
	_, _, err := dec.DecodeArray("param", sm, decodeArrayStringSchema())
	require.Error(t, err)
}
