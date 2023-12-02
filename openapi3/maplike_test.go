package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaplikeMethods(t *testing.T) {
	t.Parallel()

	t.Run("*Responses", func(t *testing.T) {
		t.Parallel()
		t.Run("nil", func(t *testing.T) {
			x := (*Responses)(nil)
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*ResponseRef{}, x.Map())
			require.Equal(t, (*ResponseRef)(nil), x.Value("key"))
			require.Panics(t, func() { x.Set("key", &ResponseRef{}) })
		})
		t.Run("nonnil", func(t *testing.T) {
			x := &Responses{}
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*ResponseRef{}, x.Map())
			require.Equal(t, (*ResponseRef)(nil), x.Value("key"))
			x.Set("key", &ResponseRef{})
			require.Equal(t, 1, x.Len())
			require.Equal(t, map[string]*ResponseRef{"key": {}}, x.Map())
			require.Equal(t, &ResponseRef{}, x.Value("key"))
		})
	})

	t.Run("*Callback", func(t *testing.T) {
		t.Parallel()
		t.Run("nil", func(t *testing.T) {
			x := (*Callback)(nil)
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*PathItem{}, x.Map())
			require.Equal(t, (*PathItem)(nil), x.Value("key"))
			require.Panics(t, func() { x.Set("key", &PathItem{}) })
		})
		t.Run("nonnil", func(t *testing.T) {
			x := &Callback{}
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*PathItem{}, x.Map())
			require.Equal(t, (*PathItem)(nil), x.Value("key"))
			x.Set("key", &PathItem{})
			require.Equal(t, 1, x.Len())
			require.Equal(t, map[string]*PathItem{"key": {}}, x.Map())
			require.Equal(t, &PathItem{}, x.Value("key"))
		})
	})

	t.Run("*Paths", func(t *testing.T) {
		t.Parallel()
		t.Run("nil", func(t *testing.T) {
			x := (*Paths)(nil)
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*PathItem{}, x.Map())
			require.Equal(t, (*PathItem)(nil), x.Value("key"))
			require.Panics(t, func() { x.Set("key", &PathItem{}) })
		})
		t.Run("nonnil", func(t *testing.T) {
			x := &Paths{}
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]*PathItem{}, x.Map())
			require.Equal(t, (*PathItem)(nil), x.Value("key"))
			x.Set("key", &PathItem{})
			require.Equal(t, 1, x.Len())
			require.Equal(t, map[string]*PathItem{"key": {}}, x.Map())
			require.Equal(t, &PathItem{}, x.Value("key"))
		})
	})

}
