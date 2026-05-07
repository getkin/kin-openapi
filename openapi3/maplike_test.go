package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestMaplikeMethods(t *testing.T) {
	t.Parallel()

	t.Run("*openapi3.Responses", func(t *testing.T) {
		t.Parallel()
		t.Run("nil", func(t *testing.T) {
			x := (*openapi3.Responses)(nil)
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.ResponseRef{}, x.Map())
			require.Equal(t, (*openapi3.ResponseRef)(nil), x.Value("key"))
			require.Panics(t, func() { x.Set("key", &openapi3.ResponseRef{}) })
			require.NotPanics(t, func() { x.Delete("key") })
		})
		t.Run("nonnil", func(t *testing.T) {
			x := &openapi3.Responses{}
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.ResponseRef{}, x.Map())
			require.Equal(t, (*openapi3.ResponseRef)(nil), x.Value("key"))
			x.Set("key", &openapi3.ResponseRef{})
			require.Equal(t, 1, x.Len())
			require.Equal(t, map[string]*openapi3.ResponseRef{"key": {}}, x.Map())
			require.Equal(t, &openapi3.ResponseRef{}, x.Value("key"))
			x.Delete("key")
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.ResponseRef{}, x.Map())
			require.Equal(t, (*openapi3.ResponseRef)(nil), x.Value("key"))
			require.NotPanics(t, func() { x.Delete("key") })
		})
	})

	t.Run("*openapi3.Callback", func(t *testing.T) {
		t.Parallel()
		t.Run("nil", func(t *testing.T) {
			x := (*openapi3.Callback)(nil)
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.PathItem{}, x.Map())
			require.Equal(t, (*openapi3.PathItem)(nil), x.Value("key"))
			require.Panics(t, func() { x.Set("key", &openapi3.PathItem{}) })
			require.NotPanics(t, func() { x.Delete("key") })
		})
		t.Run("nonnil", func(t *testing.T) {
			x := &openapi3.Callback{}
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.PathItem{}, x.Map())
			require.Equal(t, (*openapi3.PathItem)(nil), x.Value("key"))
			x.Set("key", &openapi3.PathItem{})
			require.Equal(t, 1, x.Len())
			require.Equal(t, map[string]*openapi3.PathItem{"key": {}}, x.Map())
			require.Equal(t, &openapi3.PathItem{}, x.Value("key"))
			x.Delete("key")
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.PathItem{}, x.Map())
			require.Equal(t, (*openapi3.PathItem)(nil), x.Value("key"))
			require.NotPanics(t, func() { x.Delete("key") })
		})
	})

	t.Run("*openapi3.Paths", func(t *testing.T) {
		t.Parallel()
		t.Run("nil", func(t *testing.T) {
			x := (*openapi3.Paths)(nil)
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.PathItem{}, x.Map())
			require.Equal(t, (*openapi3.PathItem)(nil), x.Value("key"))
			require.Panics(t, func() { x.Set("key", &openapi3.PathItem{}) })
			require.NotPanics(t, func() { x.Delete("key") })
		})
		t.Run("nonnil", func(t *testing.T) {
			x := &openapi3.Paths{}
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.PathItem{}, x.Map())
			require.Equal(t, (*openapi3.PathItem)(nil), x.Value("key"))
			x.Set("key", &openapi3.PathItem{})
			require.Equal(t, 1, x.Len())
			require.Equal(t, map[string]*openapi3.PathItem{"key": {}}, x.Map())
			require.Equal(t, &openapi3.PathItem{}, x.Value("key"))
			x.Delete("key")
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*openapi3.PathItem{}, x.Map())
			require.Equal(t, (*openapi3.PathItem)(nil), x.Value("key"))
			require.NotPanics(t, func() { x.Delete("key") })
		})
	})

}
