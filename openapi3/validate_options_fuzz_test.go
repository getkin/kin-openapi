package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// Turning a check off never rejects a document the default checks accept,
// and collecting every error instead of the first changes nothing about
// whether there is one.
func FuzzValidateOptions(f *testing.F) {
	seedFromTestdata(f)
	f.Add([]byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"},"paths":{},"components":{"schemas":{"a":{"type":"string","format":"email","pattern":"^a","default":"b","example":"c"}}}}`))
	f.Add([]byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"},"paths":{},"components":{"schemas":{"a":{"$ref":"#/components/schemas/b","x-a":1,"description":"d"},"b":{}}}}`))

	relaxed := map[string]openapi3.ValidationOption{
		"DisableSchemaFormatValidation":   openapi3.DisableSchemaFormatValidation(),
		"DisableSchemaPatternValidation":  openapi3.DisableSchemaPatternValidation(),
		"DisableSchemaDefaultsValidation": openapi3.DisableSchemaDefaultsValidation(),
		"DisableExamplesValidation":       openapi3.DisableExamplesValidation(),
		"AllowExtensionsWithRef":          openapi3.AllowExtensionsWithRef(),
		"AllowExtraSiblingFields":         openapi3.AllowExtraSiblingFields("description", "summary", "type"),
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxFuzzInput {
			t.Skip()
		}
		doc, err := newFuzzLoader().LoadFromData(data)
		if err != nil {
			return
		}

		strict := doc.Validate(t.Context())
		multi := doc.Validate(t.Context(), openapi3.EnableMultiError())
		require.Equal(t, strict == nil, multi == nil, "first: %v\nall: %v", strict, multi)
		if strict != nil {
			return
		}
		for name, option := range relaxed {
			require.NoError(t, doc.Validate(t.Context(), option), name)
		}
	})
}
