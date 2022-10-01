package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue492(t *testing.T) {
	spec := []byte(`components:
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
`)

	s, err := NewLoader().LoadFromData(spec)
	require.NoError(t, err)

	// verify that the expected format works
	err = s.Components.Schemas["Server"].Value.VisitJSON(map[string]interface{}{
		"name": "kin-openapi",
		"time": "2001-02-03T04:05:06.789Z",
	})
	require.NoError(t, err)

	// verify that the issue is fixed
	err = s.Components.Schemas["Server"].Value.VisitJSON(map[string]interface{}{
		"name": "kin-openapi",
		"time": "2001-02-03T04:05:06:789Z",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Error at \"/time\": string doesn't match the format \"date-time\"")
}
