package openapi3

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue430(t *testing.T) {
	schema := NewOneOfSchema(
		NewStringSchema().WithFormat("ipv4"),
		NewStringSchema().WithFormat("ipv6"),
	)

	delete(SchemaStringFormats, "ipv4")
	delete(SchemaStringFormats, "ipv6")

	err := schema.Validate(context.Background())
	require.NoError(t, err)

	data := map[string]bool{
		"127.0.1.1": true,

		// https://stackoverflow.com/a/48519490/1418165

		// v4
		"192.168.0.1": true,
		// "192.168.0.1:80" doesn't parse per net.ParseIP()

		// v6
		"::FFFF:C0A8:1":                        false,
		"::FFFF:C0A8:0001":                     false,
		"0000:0000:0000:0000:0000:FFFF:C0A8:1": false,
		// "::FFFF:C0A8:1%1" doesn't parse per net.ParseIP()
		"::FFFF:192.168.0.1": false,
		// "[::FFFF:C0A8:1]:80" doesn't parse per net.ParseIP()
		// "[::FFFF:C0A8:1%1]:80" doesn't parse per net.ParseIP()
		"2001:db8::": false,
	}

	for datum := range data {
		err = schema.VisitJSON(datum)
		require.Error(t, err, ErrOneOfConflict.Error())
	}

	DefineIPv4Format()
	DefineIPv6Format()

	ipv4Validator := NewIPValidator(true)
	ipv6Validator := NewIPValidator(false)

	for datum, isV4 := range data {
		err = schema.VisitJSON(datum)
		require.NoError(t, err)
		if isV4 {
			assert.Nil(t, ipv4Validator.Validate(datum), "%q should be IPv4", datum)
			assert.NotNil(t, ipv6Validator.Validate(datum), "%q should not be IPv6", datum)
		} else {
			assert.NotNil(t, ipv4Validator.Validate(datum), "%q should not be IPv4", datum)
			assert.Nil(t, ipv6Validator.Validate(datum), "%q should be IPv6", datum)
		}
	}
}

func TestFormatCallback_WrapError(t *testing.T) {
	var errSomething = errors.New("something error")

	DefineStringFormatValidator("foobar", NewCallbackValidator(func(value string) error {
		return errSomething
	}))

	s := &Schema{Format: "foobar"}
	err := s.VisitJSONString("blablabla")

	assert.ErrorIs(t, err, errSomething)

	delete(SchemaStringFormats, "foobar")
}

func TestReversePathInMessageSchemaError(t *testing.T) {
	DefineIPv4Format()

	SchemaErrorDetailsDisabled = true

	const spc = `
components:
  schemas:
    Something:
      type: object
      properties:
        ip:
          type: string
          format: ipv4
`
	l := NewLoader()

	doc, err := l.LoadFromData([]byte(spc))
	require.NoError(t, err)

	err = doc.Components.Schemas["Something"].Value.VisitJSON(map[string]any{
		`ip`: `123.0.0.11111`,
	})

	require.EqualError(t, err, `Error at "/ip": string doesn't match the format "ipv4": Not an IP address`)

	delete(SchemaStringFormats, "ipv4")
	SchemaErrorDetailsDisabled = false
}

func TestUuidFormat(t *testing.T) {

	type testCase struct {
		name    string
		value   string
		wantErr bool
	}

	DefineStringFormatValidator("uuid", NewRegexpFormatValidator(FormatOfStringForUUIDOfRFC4122))
	testCases := []testCase{
		{
			name:    "invalid",
			value:   "foo",
			wantErr: true,
		},
		{
			name:    "uuid v1",
			value:   "77e66540-ca29-11ed-afa1-0242ac120002",
			wantErr: false,
		},
		{
			name:    "uuid v4",
			value:   "00f4d301-b9f4-4366-8907-2b5a03430aa1",
			wantErr: false,
		},
		{
			name:    "uuid nil",
			value:   "00000000-0000-0000-0000-000000000000",
			wantErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewUUIDSchema().VisitJSON(tc.value)
			var schemaError = &SchemaError{}
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &schemaError)

				require.NotZero(t, schemaError.Reason)
				require.NotContains(t, schemaError.Reason, fmt.Sprint(tc.value))
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestNumberFormats(t *testing.T) {
	type testCase struct {
		name    string
		typ     string
		format  string
		value   any
		wantErr bool
	}
	DefineNumberFormatValidator("lessThan10", NewCallbackValidator(func(value float64) error {
		if value >= 10 {
			return errors.New("not less than 10")
		}
		return nil
	}))
	DefineIntegerFormatValidator("odd", NewCallbackValidator(func(value int64) error {
		if value%2 == 0 {
			return errors.New("not odd")
		}
		return nil
	}))
	testCases := []testCase{
		{
			name:    "invalid number",
			value:   "test",
			typ:     "number",
			format:  "",
			wantErr: true,
		},
		{
			name:    "zero float64",
			value:   0.0,
			typ:     "number",
			format:  "lessThan10",
			wantErr: false,
		},
		{
			name:    "11",
			value:   11.0,
			typ:     "number",
			format:  "lessThan10",
			wantErr: true,
		},
		{
			name:    "odd 11",
			value:   11.0,
			typ:     "integer",
			format:  "odd",
			wantErr: false,
		},
		{
			name:    "even 12",
			value:   12.0,
			typ:     "integer",
			format:  "odd",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema := &Schema{
				Type:   &Types{tc.typ},
				Format: tc.format,
			}
			err := schema.VisitJSON(tc.value)
			var schemaError = &SchemaError{}
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &schemaError)

				require.NotZero(t, schemaError.Reason)
				require.NotContains(t, schemaError.Reason, fmt.Sprint(tc.value))
			} else {
				require.Nil(t, err)
			}
		})
	}
}
