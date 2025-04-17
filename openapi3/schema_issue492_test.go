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
	require.EqualError(t, err, "Error at \"/time\": string doesn't match the format \"date-time\" (regular expression \"^[0-9]{4}-(0[0-9]|10|11|12)-([0-2][0-9]|30|31)T[0-9]{2}:[0-9]{2}:[0-9]{2}(\\.[0-9]+)?(Z|(\\+|-)[0-9]{2}:[0-9]{2})?$\")\nSchema:\n  {\n    \"format\": \"date-time\",\n    \"type\": \"string\"\n  }\n\nValue:\n  \"2001-02-03T04:05:06:789Z\"\n")
}
