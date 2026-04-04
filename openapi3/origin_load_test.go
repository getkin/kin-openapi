package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOrigin_LoadAllTestdata verifies that enabling origin tracking does not
// break loading of any spec in the testdata directory. It catches regressions
// where __origin__ leaks into fields and causes unmarshal failures or panics.
func TestOrigin_LoadAllTestdata(t *testing.T) {
	specs := []struct {
		name         string
		file         string
		externalRefs bool
	}{
		{name: "spec", file: "testdata/spec.yaml", externalRefs: true},
		{name: "main", file: "testdata/main.yaml", externalRefs: true},
		{name: "link-example", file: "testdata/link-example.yaml"},
		{name: "lxkns", file: "testdata/lxkns.yaml"},
		{name: "303bis", file: "testdata/303bis/service.yaml", externalRefs: true},
		{name: "issue638/test1", file: "testdata/issue638/test1.yaml", externalRefs: true},
		{name: "issue638/test2", file: "testdata/issue638/test2.yaml", externalRefs: true},
		{name: "refInRefInProperty", file: "testdata/refInRefInProperty/openapi.yaml", externalRefs: true},
		{name: "circularRef2", file: "testdata/circularRef2/circular2.yaml", externalRefs: true},
		{name: "origin/simple", file: "testdata/origin/simple.yaml"},
		{name: "origin/parameters", file: "testdata/origin/parameters.yaml"},
		{name: "origin/security", file: "testdata/origin/security.yaml"},
		{name: "origin/components", file: "testdata/origin/components.yaml"},
		{name: "origin/extensions", file: "testdata/origin/extensions.yaml"},
		{name: "origin/alias", file: "testdata/origin/alias.yaml"},
		{name: "origin/any_fields", file: "testdata/origin/any_fields.yaml"},
		{name: "origin/required_sequence", file: "testdata/origin/required_sequence.yaml"},
		{name: "origin/headers", file: "testdata/origin/headers.yaml"},
		{name: "origin/example", file: "testdata/origin/example.yaml"},
		{name: "origin/example_with_array", file: "testdata/origin/example_with_array.yaml"},
		{name: "origin/external", file: "testdata/origin/external.yaml", externalRefs: true},
		{name: "origin/additional_properties", file: "testdata/origin/additional_properties.yaml"},
		{name: "origin/request_body", file: "testdata/origin/request_body.yaml"},
		{name: "origin/xml", file: "testdata/origin/xml.yaml"},
		{name: "origin/external_docs", file: "testdata/origin/external_docs.yaml"},
	}

	for _, tc := range specs {
		t.Run(tc.name, func(t *testing.T) {
			loader := NewLoader()
			loader.IncludeOrigin = true
			loader.IsExternalRefsAllowed = tc.externalRefs
			loader.Context = context.Background()

			_, err := loader.LoadFromFile(tc.file)
			require.NoError(t, err)
		})
	}
}
