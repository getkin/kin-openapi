package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
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
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name": "kin-openapi",
		"date": "2001-00-03",
	})
	require.EqualError(t, err, `Error at "/date": string doesn't match the format "date": string doesn't match pattern "`+openapi3.FormatOfStringDate+`"`)
}

func TestDateZeroDay(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name": "kin-openapi",
		"date": "2001-02-00",
	})
	require.EqualError(t, err, `Error at "/date": string doesn't match the format "date": string doesn't match pattern "`+openapi3.FormatOfStringDate+`"`)
}

func TestDateTimeZeroMonth(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2001-00-03T04:05:06.789Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+openapi3.FormatOfStringDateTime+`"`)
}

func TestDateTimeZeroDay(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2001-02-00T04:05:06.789Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+openapi3.FormatOfStringDateTime+`"`)
}

func TestDateTimeLeapSecond(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
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
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2016-12-31T24:00:00.000Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+openapi3.FormatOfStringDateTime+`"`)
}

func TestDateTimeMinuteOutOfBounds(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2016-12-31T23:60:00.000Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+openapi3.FormatOfStringDateTime+`"`)
}

func TestDateTimeSecondOutOfBounds(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2016-12-31T23:59:61.000Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+openapi3.FormatOfStringDateTime+`"`)
}

func TestDateDayOutOfMonth(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	for _, date := range []string{"2023-01-31", "2023-02-28", "2024-02-29", "2000-02-29", "2023-04-30", "2023-12-31"} {
		err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
			"name": "kin-openapi",
			"date": date,
		})
		require.NoError(t, err, date)
	}

	for _, date := range []string{"2023-02-29", "2100-02-29", "2024-02-30", "2024-02-31", "2023-04-31", "2023-06-31", "2023-09-31", "2023-11-31"} {
		err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
			"name": "kin-openapi",
			"date": date,
		})
		require.EqualError(t, err, `Error at "/date": string doesn't match the format "date": parsing time "`+date+`": day out of range`)
	}
}

func TestDateTimeDayOutOfMonth(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2024-02-29T04:05:06.789Z",
	})
	require.NoError(t, err)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2023-02-29T04:05:06.789Z",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": parsing time "2023-02-29": day out of range`)

	err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
		"name":     "kin-openapi",
		"datetime": "2023-04-31T04:05:06+02:00",
	})
	require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": parsing time "2023-04-31": day out of range`)
}

func TestDateTimeOffsetOutOfBounds(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(DateTimeSpec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	for _, datetime := range []string{"2016-12-31T23:59:59+00:00", "2016-12-31T23:59:59-23:59", "2016-12-31T23:59:59+14:00"} {
		err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
			"name":     "kin-openapi",
			"datetime": datetime,
		})
		require.NoError(t, err, datetime)
	}

	for _, datetime := range []string{"2016-12-31T23:59:59+24:00", "2016-12-31T23:59:59-99:00", "2016-12-31T23:59:59+01:60", "2016-12-31T23:59:59+99:99"} {
		err = doc.Components.Schemas["Server"].Value.VisitJSON(map[string]any{
			"name":     "kin-openapi",
			"datetime": datetime,
		})
		require.EqualError(t, err, `Error at "/datetime": string doesn't match the format "date-time": string doesn't match pattern "`+openapi3.FormatOfStringDateTime+`"`)
	}
}

func TestDateDayOutOfMonthInExample(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"3.0.3", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte(`
openapi: ` + version + `
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Day:
      type: string
      format: date
      example: "2023-02-29"
`))
			require.NoError(t, err)

			err = doc.Validate(loader.Context)
			require.ErrorContains(t, err, `schema "Day": invalid example`)
			require.ErrorContains(t, err, `parsing time "2023-02-29": day out of range`)
		})
	}
}
