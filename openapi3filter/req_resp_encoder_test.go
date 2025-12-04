package openapi3filter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterAndUnregisterBodyEncoder(t *testing.T) {
	var encoder BodyEncoder
	encoder = func(body any) (data []byte, err error) {
		return []byte(strings.Join(body.([]string), ",")), nil
	}
	const contentType = "text/csv"

	originalEncoder := RegisteredBodyEncoder(contentType)
	require.Nil(t, originalEncoder)

	RegisterBodyEncoder(contentType, encoder)
	require.Equal(t, fmt.Sprint(encoder), fmt.Sprint(RegisteredBodyEncoder(contentType)))

	body := []string{"foo", "bar"}
	got, err := encodeBody(body, contentType)

	require.NoError(t, err)
	require.Equal(t, []byte("foo,bar"), got)

	UnregisterBodyEncoder(contentType)

	originalEncoder = RegisteredBodyEncoder(contentType)
	require.Nil(t, originalEncoder)

	_, err = encodeBody(body, contentType)
	require.Equal(t, &ParseError{
		Kind:   KindUnsupportedFormat,
		Reason: prefixUnsupportedCT + ` "text/csv"`,
	}, err)
}
