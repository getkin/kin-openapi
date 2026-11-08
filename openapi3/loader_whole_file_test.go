package openapi3

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoaderWholeFileSchemaReuse(t *testing.T) {
	t.Parallel()
	const layers = 14
	files := map[string][]byte{
		"openapi.yaml": []byte(`{openapi: 3.0.3, info: {title: t, version: '1'}, paths: {}, components: {schemas: {Root: {$ref: './A0.yaml'}}}}`),
	}
	for i := 0; i <= layers; i++ {
		for _, prefix := range []string{"A", "B"} {
			body := "{type: object}"
			if i < layers {
				body = fmt.Sprintf(`{type: object, properties: {a: {$ref: './A%d.yaml'}, b: {$ref: './B%d.yaml'}}}`, i+1, i+1)
			}
			files[fmt.Sprintf("%s%d.yaml", prefix, i)] = []byte(body)
		}
	}
	loader := NewLoader()
	reads := 0
	loader.ReadFromURIFunc = func(_ *Loader, location *url.URL) ([]byte, error) {
		reads++
		data, ok := files[location.Path]
		if !ok {
			return nil, fmt.Errorf("unexpected URI %q", location)
		}
		return data, nil
	}
	doc, err := loader.LoadFromFile("openapi.yaml")
	require.NoError(t, err)
	require.Equal(t, 2*layers+2, reads)
	root := doc.Components.Schemas["Root"].Value
	for range layers - 1 {
		a, b := root.Properties["a"].Value, root.Properties["b"].Value
		require.Same(t, a.Properties["a"].Value, b.Properties["a"].Value)
		require.Same(t, a.Properties["b"].Value, b.Properties["b"].Value)
		root = a
	}
}

func TestLoaderWholeFileSchemaScopeAndCycles(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"left/node.yaml":  `{type: object, properties: {leaf: {$ref: './leaf.yaml'}, next: {$ref: '../right/node.yaml'}}}`,
		"left/leaf.yaml":  `{type: string}`,
		"right/node.yaml": `{type: object, properties: {leaf: {$ref: './leaf.yaml'}, next: {$ref: '../left/node.yaml'}}}`,
		"right/leaf.yaml": `{type: integer}`,
	}
	loader := NewLoader()
	reads := make(map[string]int)
	loader.ReadFromURIFunc = func(_ *Loader, location *url.URL) ([]byte, error) {
		reads[location.Path]++
		body, ok := files[location.Path]
		if !ok {
			return nil, fmt.Errorf("unexpected URI %q", location)
		}
		return []byte(body), nil
	}
	doc, err := loader.LoadFromDataWithPath([]byte(`{openapi: 3.0.3, info: {title: t, version: '1'}, paths: {}, components: {schemas: {Left: {$ref: 'left/node.yaml'}, Right: {$ref: 'right/node.yaml'}, Alias: {$ref: 'left/../right/node.yaml'}}}}`), &url.URL{Path: "openapi.yaml"})
	require.NoError(t, err)
	left, right := doc.Components.Schemas["Left"], doc.Components.Schemas["Right"]
	require.Same(t, right.Value, doc.Components.Schemas["Alias"].Value)
	require.Same(t, left.Value, right.Value.Properties["next"].Value)
	require.Same(t, right.Value, left.Value.Properties["next"].Value)
	require.True(t, left.Value.Properties["leaf"].Value.Type.Is("string"))
	require.True(t, right.Value.Properties["leaf"].Value.Type.Is("integer"))
	require.Equal(t, "left/leaf.yaml", left.Value.Properties["leaf"].RefPath().Path)
	require.Equal(t, "right/leaf.yaml", right.Value.Properties["leaf"].RefPath().Path)
	for name := range files {
		require.Equal(t, 1, reads[name], name)
	}
}

func TestLoaderWholeFileSchemaSiblingsAndReload(t *testing.T) {
	t.Parallel()
	loader := NewLoader()
	body := `{type: string}`
	reads := 0
	loader.ReadFromURIFunc = func(_ *Loader, _ *url.URL) ([]byte, error) {
		reads++
		return []byte(body), nil
	}
	data := []byte(`{openapi: 3.1.0, info: {title: t, version: '1'}, paths: {}, components: {schemas: {A: {$ref: 'schema.yaml', description: first}, B: {$ref: 'schema.yaml', description: second}, C: {$ref: 'schema.yaml'}}}}`)
	for _, typ := range []string{"string", "integer"} {
		body = fmt.Sprintf("{type: %s}", typ)
		doc, err := loader.LoadFromData(data)
		require.NoError(t, err)
		require.Equal(t, "first", doc.Components.Schemas["A"].Value.Description)
		require.Equal(t, "second", doc.Components.Schemas["B"].Value.Description)
		require.Empty(t, doc.Components.Schemas["C"].Value.Description)
		require.True(t, doc.Components.Schemas["C"].Value.Type.Is(typ))
	}
	require.Equal(t, 2, reads)
}

func TestLoaderWholeFileSchemaCacheDoesNotInheritFragmentPath(t *testing.T) {
	t.Parallel()
	loader := NewLoader()
	loader.ReadFromURIFunc = func(_ *Loader, location *url.URL) ([]byte, error) {
		switch location.Path {
		case "defs.yaml":
			return []byte(`{Foo: {$ref: 'target.yaml', maxLength: 2}}`), nil
		case "target.yaml":
			return []byte(`{type: string}`), nil
		default:
			return nil, fmt.Errorf("unexpected URI %q", location)
		}
	}
	doc, err := loader.LoadFromDataWithPath([]byte(`{openapi: 3.1.0, info: {title: t, version: '1'}, paths: {}, components: {schemas: {Alias: {$ref: 'defs.yaml#/Foo'}, Plain: {$ref: 'target.yaml'}}}}`), &url.URL{Path: "openapi.yaml"})
	require.NoError(t, err)
	require.NoError(t, doc.Components.Schemas["Plain"].Value.VisitJSON("abc"))
	refPath := copyURI(doc.Components.Schemas["Plain"].RefPath())
	doc.InternalizeRefs(t.Context(), nil)
	data, err := doc.MarshalJSON()
	require.NoError(t, err)
	reloader := NewLoader()
	doc, err = reloader.LoadFromData(data)
	require.NoError(t, err)
	require.NoError(t, doc.Components.Schemas["Plain"].Value.VisitJSON("abc"))
	require.Error(t, doc.Components.Schemas["Alias"].Value.VisitJSON("abc"))
	require.Equal(t, "target.yaml", refPath.String())
}

func TestLoaderWholeFileSchemaRepeatedRelativeRef(t *testing.T) {
	t.Parallel()
	loader := NewLoader()
	reads := 0
	loader.ReadFromURIFunc = func(_ *Loader, _ *url.URL) ([]byte, error) {
		reads++
		if reads > 3 {
			return nil, fmt.Errorf("repeated relative reference did not terminate")
		}
		return []byte(`{type: object, properties: {next: {$ref: './sub/schema.yaml'}}}`), nil
	}
	doc, err := loader.LoadFromDataWithPath([]byte(`{openapi: 3.0.3, info: {title: t, version: '1'}, paths: {}, components: {schemas: {Root: {$ref: './sub/schema.yaml'}}}}`), &url.URL{Path: "/spec/root.yaml"})
	require.NoError(t, err)
	root := doc.Components.Schemas["Root"].Value
	require.Same(t, root, root.Properties["next"].Value)
	require.Equal(t, 1, reads)
}
