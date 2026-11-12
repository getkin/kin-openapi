package openapi2conv_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/fuzzing"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// FuzzFromV3 converts an arbitrary valid OpenAPI 3 document to OpenAPI 2 and
// back to 3: neither direction panics, and a document that converts to 2
// converts back.
func FuzzFromV3(f *testing.F) {
	for _, seed := range fuzzing.Testdata(f, "../openapi3/testdata") {
		if len(seed) <= maxFuzzInput {
			f.Add(seed)
		}
	}
	f.Add([]byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"},"paths":{"/x/{id}":{"get":{"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string"}}],"responses":{"200":{"description":"ok"}}}}}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxFuzzInput {
			t.Skip()
		}
		loader := openapi3.NewLoader()
		doc3, err := loader.LoadFromData(data)
		if err != nil || doc3.Validate(t.Context()) != nil {
			t.Skip()
		}
		doc2, err := openapi2conv.FromV3(doc3)
		if err != nil {
			return
		}
		require.NotNil(t, doc2)
		_, err = openapi2conv.ToV3(doc2)
		require.NoError(t, err)
		err = doc3.Validate(t.Context())
		require.NoError(t, err)
	})
}
