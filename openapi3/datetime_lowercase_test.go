package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// RFC 3339 section 5.6 allows the "T" and "Z" separators to alternatively be
// lowercase "t" and "z", and JSON Schema Validation section 7.3.1 defines the
// "date-time" format via RFC 3339, so both cases must validate.
func TestDateTimeLowercaseSeparators(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	for _, datetime := range []string{
		"2001-02-03t04:05:06.789z",
		"2001-02-03T04:05:06.789z",
		"2001-02-03t04:05:06.789Z",
		"2001-02-03t04:05:06Z",
		"2001-02-03t04:05:06+01:00",
	} {
		err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
			"name":     "kin-openapi",
			"datetime": datetime,
		})
		require.NoError(t, err, "value %q should validate", datetime)
	}
}
