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
	}

	for datum := range data {
		err = schema.VisitJSON(datum)
		require.Error(t, err, ErrOneOfConflict.Error())
	}

	DefineIPv4Format()
	DefineIPv6Format()

	for datum, isV4 := range data {
		err = schema.VisitJSON(datum)
		require.NoError(t, err)
		if isV4 {
			require.Nil(t, validateIPv4(datum), "%q should be IPv4", datum)
			require.NotNil(t, validateIPv6(datum), "%q should not be IPv6", datum)
		} else {
			require.NotNil(t, validateIPv4(datum), "%q should not be IPv4", datum)
			require.Nil(t, validateIPv6(datum), "%q should be IPv6", datum)
		}
	}
}

func TestFormatCallback_WrapError(t *testing.T) {
	var errSomething = errors.New("something error")

	DefineStringFormatCallback("foobar", func(value string) error {
		return errSomething
	})

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

	err = doc.Components.Schemas["Something"].Value.VisitJSON(map[string]interface{}{
		`ip`: `123.0.0.11111`,
	})

	require.EqualError(t, err, `Error at "/ip": Not an IP address`)

	delete(SchemaStringFormats, "ipv4")
	SchemaErrorDetailsDisabled = false
}

func TestUuidFormat(t *testing.T) {

	type testCase struct {
		name    string
		value   string
		wantErr bool
	}

	DefineStringFormat("uuid", FormatOfStringForUUIDOfRFC4122)
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
