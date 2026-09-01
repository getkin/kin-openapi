package openapi3filter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterAndUnregisterBodyEncoder(t *testing.T) {
	var encoder BodyEncoder = func(body any) (data []byte, err error) {
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

func TestGetBodyEncoderCaseVariantsAreDeterministic(t *testing.T) {
	var encoder BodyEncoder = func(body any) ([]byte, error) { return nil, nil }
	for _, contentType := range []string{"TEXT/CSV", "text/CSV", "text/csv"} {
		RegisterBodyEncoder(contentType, encoder)
		defer UnregisterBodyEncoder(contentType)
	}

	// None of the registrations match verbatim, so the case-insensitive search
	// decides between them: the lexicographically smallest one wins.
	got, ok := getBodyEncoder("Text/Csv")
	require.True(t, ok)
	require.Equal(t, fmt.Sprint(RegisteredBodyEncoder("TEXT/CSV")), fmt.Sprint(got))
}

func TestEncodeBodyCaseInsensitiveMediaType(t *testing.T) {
	got, err := encodeBody(map[string]any{"name": "default"}, "APPLICATION/JSON")
	require.NoError(t, err)
	require.JSONEq(t, `{"name":"default"}`, string(got))

	_, err = encodeBody(map[string]any{}, "APPLICATION")
	require.Equal(t, &ParseError{
		Kind:   KindUnsupportedFormat,
		Reason: prefixUnsupportedCT + ` "APPLICATION"`,
	}, err)
}
