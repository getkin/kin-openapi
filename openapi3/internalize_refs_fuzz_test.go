package openapi3_test

import (
	"encoding/json"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/fuzzing"
	"github.com/getkin/kin-openapi/openapi3"
)

// fuzzRootURL is where the root document of FuzzInternalizeRefs lives. A ref
// naming its file reads the root again, any other file reads the second input.
var fuzzRootURL = &url.URL{Path: "/spec/root.yaml"}

func newTwoDocLoader(root, other []byte) *openapi3.Loader {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(_ *openapi3.Loader, location *url.URL) ([]byte, error) {
		if path.Base(location.Path) == path.Base(fuzzRootURL.Path) {
			return root, nil
		}
		return other, nil
	}
	return loader
}

// internalizeSeeds pairs each root of TestInternalizeRefs with a document its
// external refs point to.
var internalizeSeeds = [][2]string{
	{"testdata/testref.openapi.yml", "testdata/components.openapi.yml"},
	{"testdata/recursiveRef/openapi.yml", "testdata/recursiveRef/components/Foo.yml"},
	{"testdata/interalizationNameCollision/api.yml", "testdata/interalizationNameCollision/schemas/book/record.yml"},
	{"testdata/issue1205/openapi.yml", "testdata/issue1205/ext.yml"},
	{"testdata/issue1205/mutual.yml", "testdata/issue1205/mutual-a.yml"},
	{"testdata/issue1205/mutual.yml", "testdata/issue1205/mutual-b.yml"},
}

// Once internalized, a document refers to nothing outside itself, loads
// without external refs allowed, and internalizing it again changes nothing.
// A valid document stays valid.
func FuzzInternalizeRefs(f *testing.F) {
	for _, pair := range internalizeSeeds {
		root, err := os.ReadFile(pair[0])
		require.NoError(f, err)
		other, err := os.ReadFile(pair[1])
		require.NoError(f, err)
		f.Add(root, other)
	}
	// Any document whose refs, whatever file they name, read itself back.
	for _, seed := range fuzzing.Testdata(f, "testdata") {
		f.Add(seed, seed)
	}
	f.Add(
		[]byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"},"paths":{"/a":{"$ref":"other.yaml#/paths/~1b"}}}`),
		[]byte(`{"paths":{"/b":{"get":{"responses":{"200":{"$ref":"#/components/responses/R"}}}}},"components":{"responses":{"R":{"description":"d","content":{"application/json":{"schema":{"$ref":"#/components/schemas/S"}}}}},"schemas":{"S":{"type":"string"}}}}`),
	)
	f.Add(
		[]byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"},"paths":{},"components":{"schemas":{"A":{"$ref":"x.json#/components/schemas/A"},"x_A":{"type":"integer"}}}}`),
		[]byte(`{"components":{"schemas":{"A":{"type":"object","properties":{"self":{"$ref":"#/components/schemas/A"},"root":{"$ref":"root.yaml#/components/schemas/x_A"}}}}}}`),
	)

	f.Fuzz(func(t *testing.T, root, other []byte) {
		if len(root)+len(other) > maxFuzzInput {
			t.Skip()
		}

		doc, err := newTwoDocLoader(root, other).LoadFromDataWithPath(root, fuzzRootURL)
		if err != nil {
			return
		}
		valid := doc.Validate(t.Context())

		doc.InternalizeRefs(t.Context(), nil)

		unfollowed := false
		forEachRef(doc, func(site refSite) {
			if site.Mapping && !strings.Contains(site.Ref, "/") {
				return // a plain schema name, not a reference
			}
			if site.Ref == "" || strings.HasPrefix(site.Ref, "#") {
				return
			}
			// The loader did not follow it, so there is nothing to internalize.
			require.True(t, site.Unfollowed, "external ref left: %q", site.Ref)
			unfollowed = true
		})

		once, err := doc.MarshalJSON()
		require.NoError(t, err)

		strict := openapi3.NewLoader()
		reloaded, err := strict.LoadFromData(once)
		if unfollowed {
			return // still refers elsewhere, which strict refuses
		}
		require.NoError(t, err, "\nonce: %s\nroot:\n%s\nother:\n%s", once, root, other)

		if valid == nil {
			require.NoError(t, reloaded.Validate(t.Context()), "\nonce: %s", once)
		}

		reloaded.InternalizeRefs(t.Context(), nil)
		twice, err := reloaded.MarshalJSON()
		require.NoError(t, err)
		require.JSONEq(t, string(once), string(twice))
	})
}

var mappingRefType = reflect.TypeFor[openapi3.MappingRef]()

// refSite is the Ref field of a reference object, path item or discriminator
// mapping.
type refSite struct {
	Ref string
	// Mapping tells a discriminator mapping value, which may be a plain name.
	Mapping bool
	// Unfollowed tells a ref the loader left without a target.
	Unfollowed bool
}

// forEachRef calls fn on every refSite reachable from v. A resolved document
// is a graph, so each pointer is visited once. Values typed any, such as
// examples and extensions, are not entered: a "$ref" key there is data.
func forEachRef(v any, fn func(refSite)) {
	seen := make(map[uintptr]struct{})
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Pointer:
			if v.IsNil() {
				return
			}
			if _, ok := seen[v.Pointer()]; ok {
				return
			}
			seen[v.Pointer()] = struct{}{}
			walk(v.Elem())
		case reflect.Struct:
			if f, ok := v.Type().FieldByName("Ref"); ok && f.Type.Kind() == reflect.String && len(f.Index) == 1 {
				site := refSite{Ref: v.FieldByIndex(f.Index).String(), Mapping: v.Type() == mappingRefType}
				if p := v.FieldByName("refPath"); p.IsValid() {
					site.Unfollowed = site.Ref != "" && p.IsNil()
				}
				fn(site)
			}
			for i := range v.NumField() {
				walk(v.Field(i))
			}
		case reflect.Slice, reflect.Array:
			for i := range v.Len() {
				walk(v.Index(i))
			}
		case reflect.Map:
			for it := v.MapRange(); it.Next(); {
				walk(it.Value())
			}
		}
	}
	walk(reflect.ValueOf(v))
}

// Internalizing changes where schemas live, not what they mean: each schema
// of the root's components accepts the same values before and after.
func FuzzInternalizeRefsKeepsMeaning(f *testing.F) {
	for _, pair := range internalizeSeeds {
		root, err := os.ReadFile(pair[0])
		require.NoError(f, err)
		other, err := os.ReadFile(pair[1])
		require.NoError(f, err)
		for _, value := range []string{`{}`, `null`, `"s"`, `1`, `[{"cat":{}}]`, `{"foo":{"bar":"x"},"cat":{"cat":1}}`} {
			f.Add(root, other, []byte(value))
		}
	}
	f.Add(
		[]byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"},"paths":{},"components":{"schemas":{"A":{"type":"object","properties":{"b":{"$ref":"other.yaml#/components/schemas/B"}}}}}}`),
		[]byte(`{"components":{"schemas":{"B":{"type":"array","items":{"$ref":"#/components/schemas/C"}},"C":{"type":"integer","maximum":3}}}}`),
		[]byte(`{"b":[1,2,4]}`),
	)

	f.Fuzz(func(t *testing.T, root, other, value []byte) {
		if len(root)+len(other)+len(value) > maxFuzzInput {
			t.Skip()
		}
		var v any
		if json.Unmarshal(value, &v) != nil {
			t.Skip()
		}

		doc, err := newTwoDocLoader(root, other).LoadFromDataWithPath(root, fuzzRootURL)
		if err != nil || doc.Components == nil {
			return
		}
		if doc.Validate(t.Context()) != nil {
			return
		}
		before := make(map[string]bool, len(doc.Components.Schemas))
		for name, schema := range doc.Components.Schemas {
			if schema != nil && schema.Value != nil {
				before[name] = schema.Value.VisitJSON(v) == nil
			}
		}

		doc.InternalizeRefs(t.Context(), nil)
		once, err := doc.MarshalJSON()
		require.NoError(t, err)
		loader := openapi3.NewLoader()
		reloaded, err := loader.LoadFromData(once)
		if err != nil {
			return // FuzzInternalizeRefs covers this
		}
		for name, accepted := range before {
			schema := reloaded.Components.Schemas[name]
			require.NotNil(t, schema, name)
			require.NotNil(t, schema.Value, name)
			require.Equal(t, accepted, schema.Value.VisitJSON(v) == nil, "schema %q, value %s\nonce: %s", name, value, once)
		}
	})
}

// Whatever names a resolver hands out, colliding, empty or invalid ones
// included, internalizing does not panic and leaves only local refs behind.
func FuzzInternalizeRefsCustomResolver(f *testing.F) {
	for _, pair := range internalizeSeeds {
		root, err := os.ReadFile(pair[0])
		require.NoError(f, err)
		other, err := os.ReadFile(pair[1])
		require.NoError(f, err)
		f.Add(root, other, "a,a,b,,#/x,~1")
	}

	f.Fuzz(func(t *testing.T, root, other []byte, names string) {
		if len(root)+len(other)+len(names) > maxFuzzInput {
			t.Skip()
		}
		doc, err := newTwoDocLoader(root, other).LoadFromDataWithPath(root, fuzzRootURL)
		if err != nil {
			return
		}
		given := strings.Split(names, ",")
		calls := 0
		doc.InternalizeRefs(t.Context(), func(*openapi3.T, openapi3.ComponentRef) string {
			name := given[calls%len(given)]
			calls++
			return name
		})
		forEachRef(doc, func(site refSite) {
			if site.Mapping && !strings.Contains(site.Ref, "/") || site.Unfollowed {
				return
			}
			require.True(t, site.Ref == "" || strings.HasPrefix(site.Ref, "#"), "external ref left: %q", site.Ref)
		})
		_, err = doc.MarshalJSON()
		require.NoError(t, err)
	})
}
