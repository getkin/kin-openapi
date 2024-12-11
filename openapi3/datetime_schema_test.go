package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var DateSpec = []byte(`
components:
  schemas:
    Server:
      properties:
        date:
          $ref: "#/components/schemas/timestamp"
        name:
          type: string
      type: object
    timestamp:
      type: string
      format: date
openapi: "3.0.1"
paths: {}
info:
  version: 1.1.1
  title: title
`[1:])

var DateTimeSpec = []byte(`
components:
  schemas:
    Server:
      properties:
        datetime:
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

func TestDateZeroMonth(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData(DateSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name": "kin-openapi",
		"date": "2001-00-03",
	})
	require.ErrorContains(t, err, `Error at "/date": string doesn't match the format "date": string doesn't match pattern "`+FormatOfStringDate+`"`)
}

func TestDateZeroDay(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData(DateSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name": "kin-openapi",
		"date": "2001-02-00",
	})
	require.ErrorContains(t, err, `Error at "/date": string doesn't match the format "date": string doesn't match pattern "`+FormatOfStringDate+`"`)
}

func TestDateTimeZeroMonth(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2001-00-03T04:05:06.789Z",
	})
	require.ErrorContains(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+FormatOfStringDateTime+`"`)
}

func TestDateTimeZeroDay(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2001-02-00T04:05:06.789Z",
	})
	require.ErrorContains(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+FormatOfStringDateTime+`"`)
}
