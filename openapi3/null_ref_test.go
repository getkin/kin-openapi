package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// A null where a component belongs decodes to a ref with neither $ref nor a
// value. The loader rejects it, but a document unmarshaled directly holds it,
// and it must marshal back to null rather than dereference the missing value.
func TestRefs_NullValueMarshalsAsNull(t *testing.T) {
	spec := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "t", "version": "1"},
		"paths": {"/a": {"get": {"parameters": [null], "responses": {"200": null}}}},
		"components": {
			"schemas": {"A": null},
			"responses": {"R": null},
			"parameters": {"P": null}
		}
	}`)
	var doc openapi3.T
	require.NoError(t, json.Unmarshal(spec, &doc))

	once, err := json.Marshal(&doc)
	require.NoError(t, err)

	var reloaded openapi3.T
	require.NoError(t, json.Unmarshal(once, &reloaded))
	twice, err := json.Marshal(&reloaded)
	require.NoError(t, err)
	require.JSONEq(t, string(once), string(twice))
}
