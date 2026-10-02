package openapi3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue1264SchemaRefSiblings(t *testing.T) {
	for _, version := range []string{"3.0.0", "3.0.3", "3.1.0", "3.2.0"} {
		t.Run(version, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData(fmt.Appendf(nil, `
openapi: %s
info:
  version: 0.0.0
  title: Reference siblings
paths: {}
components:
  schemas:
    Schema1:
      type: object
      properties:
        p:
          $ref: '#/components/schemas/Schema2'
          description: Sibling description
          example: sibling
          minLength: 3
    Schema2:
      type: string
      description: Target description
`, version))
			require.NoError(t, err)
			require.NoError(t, doc.Validate(loader.Context))
			value := doc.Components.Schemas["Schema1"].Value.Properties["p"].Value
			if doc.IsOpenAPI30() {
				require.Equal(t, "Target description", value.Description)
				require.Nil(t, value.Example)
				require.NoError(t, value.VisitJSON("a"))
			} else {
				require.Equal(t, "Sibling description", value.Description)
				require.Equal(t, "sibling", value.Example)
				require.Error(t, value.VisitJSON("a"))
			}
			require.Equal(t, "Target description", doc.Components.Schemas["Schema2"].Value.Description)
		})
	}
}

func TestIssue1264IgnoredSiblingsDoNotHideErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema string
		want   string
	}{
		{"ignored unknown field", `{"$ref":"#/components/schemas/Target","unknown":true}`, ""},
		{"ignored invalid constraint", `{"$ref":"#/components/schemas/Target","type":"invalid"}`, ""},
		{"invalid target", `{"$ref":"#/components/schemas/Invalid","description":"ignored"}`, "unsupported 'type' value"},
		{"invalid inline schema", `{"type":"invalid"}`, "unsupported 'type' value"},
		{"unknown inline field", `{"type":"string","unknown":true}`, "extra sibling fields"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData(fmt.Appendf(nil, `{
  "openapi":"3.0.3", "info":{"title":"Reference validation","version":"1"},
  "paths":{}, "components":{"schemas":{
    "Subject":%s, "Target":{"type":"string"}, "Invalid":{"type":"invalid"}
  }}
}`, tc.schema))
			require.NoError(t, err)
			// Validate the referring schema so the unused invalid component
			// does not itself become the reason for failure.
			err = doc.Components.Schemas["Subject"].Validate(loader.Context)
			if tc.want == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.want)
			}
		})
	}
}
