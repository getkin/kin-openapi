package openapi3

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func cacheTestSchema() *Schema {
	maxLen := uint64(64)
	return &Schema{
		Type: &Types{"object"},
		Properties: Schemas{
			"name": &SchemaRef{Value: &Schema{Type: &Types{"string"}, MinLength: 1, MaxLength: &maxLen}},
			"id":   &SchemaRef{Value: &Schema{Type: &Types{"string"}, Format: "even"}},
		},
		Required: []string{"name"},
	}
}

func TestJSONSchema2020CacheReusesCompiledValidator(t *testing.T) {
	schema := cacheTestSchema()
	cache := &JSONSchema2020Cache{}
	opts := []SchemaValidationOption{EnableJSONSchema2020(), WithJSONSchema2020Cache(cache)}

	require.NoError(t, schema.VisitJSON(map[string]any{"name": "a"}, opts...))
	first, ok := cache.validators.Load(schema)
	require.True(t, ok)
	require.NotNil(t, first)

	require.NoError(t, schema.VisitJSON(map[string]any{"name": "b"}, opts...))
	second, _ := cache.validators.Load(schema)
	require.Same(t, first, second)

	// A cached validator reports the same errors as a fresh one.
	bad := map[string]any{"name": ""}
	uncached := schema.VisitJSON(bad, EnableJSONSchema2020())
	require.Error(t, uncached)
	want := uncached.Error()
	require.EqualError(t, schema.VisitJSON(bad, opts...), want)
}

func TestJSONSchema2020CacheKeepsBuiltInFallback(t *testing.T) {
	target := &Schema{Type: &Types{"string"}}
	schema := &Schema{
		Type:       &Types{"object"},
		Properties: Schemas{"ref": &SchemaRef{Ref: "#/components/schemas/Target", Value: target}},
	}
	cache := &JSONSchema2020Cache{}
	opts := []SchemaValidationOption{EnableJSONSchema2020(), WithJSONSchema2020Cache(cache)}

	for range 2 {
		require.NoError(t, schema.VisitJSON(map[string]any{"ref": "x"}, opts...))
		require.Error(t, schema.VisitJSON(map[string]any{"ref": 1}, opts...))
	}
	cached, ok := cache.validators.Load(schema)
	require.True(t, ok)
	require.Nil(t, cached.(*jsonSchemaValidator))
}

func TestJSONSchema2020CacheSkippedWithPerValidationFormats(t *testing.T) {
	schema := cacheTestSchema()
	cache := &JSONSchema2020Cache{}
	even := WithStringFormatValidator("even", NewCallbackValidator(func(s string) error {
		if len(s)%2 != 0 {
			return errors.New("odd length")
		}
		return nil
	}))
	opts := []SchemaValidationOption{EnableJSONSchema2020(), WithJSONSchema2020Cache(cache)}

	require.NoError(t, schema.VisitJSON(map[string]any{"name": "a", "id": "abc"}, opts...))
	require.Error(t, schema.VisitJSON(map[string]any{"name": "a", "id": "abc"}, append(opts, even)...))
	require.NoError(t, schema.VisitJSON(map[string]any{"name": "a", "id": "abc"}, opts...))
}

func BenchmarkJSONSchema2020Cache(b *testing.B) {
	value := map[string]any{"name": "Jane", "id": "ab"}
	b.Run("uncached", func(b *testing.B) {
		schema := cacheTestSchema()
		for b.Loop() {
			if err := schema.VisitJSON(value, EnableJSONSchema2020()); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("cached", func(b *testing.B) {
		schema := cacheTestSchema()
		opts := []SchemaValidationOption{EnableJSONSchema2020(), WithJSONSchema2020Cache(&JSONSchema2020Cache{})}
		for b.Loop() {
			if err := schema.VisitJSON(value, opts...); err != nil {
				b.Fatal(err)
			}
		}
	})
}
