package openapi3filter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOrderedJSONBodyDecoder(t *testing.T) {
	value, err := OrderedJSONBodyDecoder(strings.NewReader(`{"b": 1, "a": [{"z": null, "y": 2.5}], "c": "s"}`), nil, nil, nil)
	require.NoError(t, err)

	obj, ok := value.(*openapi3.OrderedObject)
	require.True(t, ok)
	require.Equal(t, []string{"b", "a", "c"}, obj.Names())

	a, _ := obj.Get("a")
	inner, ok := a.([]any)[0].(*openapi3.OrderedObject)
	require.True(t, ok)
	require.Equal(t, []string{"z", "y"}, inner.Names())

	_, err = OrderedJSONBodyDecoder(strings.NewReader(`{"a": 1`), nil, nil, nil)
	require.Error(t, err)
	_, err = OrderedJSONBodyDecoder(strings.NewReader(`{"a": 1} 2`), nil, nil, nil)
	require.Error(t, err)

	_, err = OrderedJSONBodyDecoder(strings.NewReader(strings.Repeat("[", 10000)+strings.Repeat("]", 10000)), nil, nil, nil)
	require.NoError(t, err)
	_, err = OrderedJSONBodyDecoder(strings.NewReader(strings.Repeat("[", 10001)+strings.Repeat("]", 10001)), nil, nil, nil)
	require.ErrorContains(t, err, "exceeded max depth")
}
