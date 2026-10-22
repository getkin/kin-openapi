package openapi2conv_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// maxFuzzInput bounds fuzzer inputs so the fuzzer spends its budget on
// conversion logic rather than on allocating huge documents.
const maxFuzzInput = 32 << 10

// FuzzToV3 converts an arbitrary OpenAPI 2 document to OpenAPI 3 and back.
func FuzzToV3(f *testing.F) {
	paths, err := filepath.Glob("../openapi2/testdata/*.json")
	require.NoError(f, err)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		require.NoError(f, err)
		if len(data) <= maxFuzzInput {
			f.Add(data)
		}
	}
	f.Add([]byte(`{"swagger":"2.0","info":{"title":"t","version":"1"},"paths":{}}`))
	f.Add([]byte(`{"swagger":"2.0","paths":{"/x/{id}":{"get":{"parameters":[{"name":"id","in":"path","type":"string","required":true}],"responses":{"200":{"description":"ok"}}}}}}`))
	f.Add([]byte(`{"swagger":"2.0","definitions":{"A":{"type":"object","properties":{"b":{"$ref":"#/definitions/A"}}}},"paths":{}}`))
	f.Add([]byte(`{"swagger":"2.0","consumes":["multipart/form-data"],"paths":{"/f":{"post":{"parameters":[{"name":"file","in":"formData","type":"file"}],"responses":{"200":{"description":"ok"}}}}}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxFuzzInput {
			t.Skip()
		}

		var doc2 openapi2.T
		if err := json.Unmarshal(data, &doc2); err != nil {
			t.Skip()
		}

		doc3, err := openapi2conv.ToV3(&doc2)
		if err != nil {
			return
		}
		require.NotNil(t, doc3)

		roundTrip := func(doc *openapi3.T) *openapi3.T {
			back2, err := openapi2conv.FromV3(doc)
			if err != nil {
				return nil
			}
			next3, err := openapi2conv.ToV3(back2)
			if err != nil {
				return nil
			}
			return next3
		}

		once := roundTrip(doc3)
		if once == nil {
			return
		}
		twice := roundTrip(once)
		if twice == nil {
			return
		}

		if err := once.Validate(t.Context()); err != nil {
			return
		}

		onceJSON, err := json.Marshal(once)
		require.NoError(t, err)
		twiceJSON, err := json.Marshal(twice)
		require.NoError(t, err)

		require.JSONEq(t, string(onceJSON), string(twiceJSON))
	})
}
