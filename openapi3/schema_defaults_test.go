package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestSettingDefaultsLeavesSchemaUnchanged(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Job:
      type: object
      properties:
        options:
          type: object
          default: {}
          properties:
            timeout: {type: integer, default: 30}
        steps:
          type: array
          default: [{}]
          items:
            type: object
            properties:
              retries: {type: integer, default: 3}
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	job := doc.Components.Schemas["Job"].Value

	expected := map[string]any{
		"options": map[string]any{"timeout": float64(30)},
		"steps":   []any{map[string]any{"retries": float64(3)}},
	}
	visit := func() map[string]any {
		t.Helper()
		defaultsSet := false
		value := map[string]any{}
		err := job.VisitJSON(value, openapi3.VisitAsRequest(), openapi3.DefaultsSet(func() { defaultsSet = true }))
		require.NoError(t, err)
		require.True(t, defaultsSet)
		require.Equal(t, expected, value)
		return value
	}
	requireDefaultsUnchanged := func() {
		t.Helper()
		require.Equal(t, map[string]any{}, job.Properties["options"].Value.Default)
		require.Equal(t, []any{map[string]any{}}, job.Properties["steps"].Value.Default)
	}

	first := visit()
	requireDefaultsUnchanged()

	// Values handed out must not alias the schema's defaults either.
	first["options"].(map[string]any)["extra"] = true
	first["steps"].([]any)[0].(map[string]any)["extra"] = true
	requireDefaultsUnchanged()

	// A second visit starts again from the pristine defaults.
	second := visit()
	require.NotContains(t, second["options"], "extra")
	requireDefaultsUnchanged()
}
