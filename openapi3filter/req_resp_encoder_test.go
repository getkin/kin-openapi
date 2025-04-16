package openapi3filter

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterAndUnregisterBodyEncoder(t *testing.T) {
	var encoder BodyEncoder
	encoder = func(body interface{}) (data []byte, err error) {
		return []byte(strings.Join(body.([]string), ",")), nil
	}
	contentType := "text/csv"
	h := make(http.Header)
	h.Set(headerCT, contentType)

	originalEncoder := RegisteredBodyEncoder(contentType)
	require.Nil(t, originalEncoder)

	RegisterBodyEncoder(contentType, encoder)
	require.Equal(t, fmt.Sprintf("%v", encoder), fmt.Sprintf("%v", RegisteredBodyEncoder(contentType)))

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
