package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue492(t *testing.T) {
	spec := []byte(`
components:
  schemas:
    Server:
      properties:
        time:
          $ref: "#/components/schemas/timestamp"
        name:
          type: string
      type: object
    timestamp:
      type: string
      format: date-time
openapi: "3.0.1"
paths: {}
info:
  version: 1.1.1
  title: title
`[1:])

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	// verify that the expected format works
	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name": "kin-openapi",
		"time": "2001-02-03T04:05:06.789Z",
	})
	require.NoError(t, err)

	// verify that the issue is fixed
	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name": "kin-openapi",
		"time": "2001-02-03T04:05:06:789Z",
	})
	require.ErrorContains(t, err, `Error at "/time": string doesn't match the format "date-time": string doesn't match pattern "^[0-9]{4}-(0[0-9]|10|11|12)-([0-2][0-9]|30|31)T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?(Z|(\+|-)[0-9]{2}:[0-9]{2})?$"`)
}
