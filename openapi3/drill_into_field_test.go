package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDrillIntoFieldSliceIndex(t *testing.T) {
	slice := []string{"a", "b"}

	for _, idx := range []string{"0", "1"} {
		v, err := drillIntoField(slice, idx)
		require.NoError(t, err)
		require.Equal(t, slice[idx[0]-'0'], v)
	}

	for _, idx := range []string{"2", "+1", "-0", "-1", "x", "", "99999999999999999999"} {
		_, err := drillIntoField(slice, idx)
		require.Error(t, err, idx)
	}
}
