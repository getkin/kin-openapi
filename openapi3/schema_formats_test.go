package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue430(t *testing.T) {
	schema := NewOneOfSchema(
		NewStringSchema().WithFormat("ipv4"),
		NewStringSchema().WithFormat("ipv6"),
	)

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
