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
	require.EqualError(t, err, `Error at "/date": string doesn't match the format "date": string doesn't match pattern "`+FormatOfStringDate+`"`)
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
	require.EqualError(t, err, `Error at "/date": string doesn't match the format "date": string doesn't match pattern "`+FormatOfStringDate+`"`)
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
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+FormatOfStringDateTime+`"`)
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
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+FormatOfStringDateTime+`"`)
}

func TestDateTimeLeapSecond(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2016-12-31T23:59:60.000Z", // exact time of the most recent leap second
	})
	require.NoError(t, err)
}

func TestDateTimeHourOutOfBounds(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2016-12-31T24:00:00.000Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+FormatOfStringDateTime+`"`)
}

func TestDateTimeMinuteOutOfBounds(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2016-12-31T23:60:00.000Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+FormatOfStringDateTime+`"`)
}

func TestDateTimeSecondOutOfBounds(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2016-12-31T23:59:61.000Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+FormatOfStringDateTime+`"`)
}
