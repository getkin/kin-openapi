package openapi2

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadingSwagger(t *testing.T) {
	var doc Swagger

	input, err := ioutil.ReadFile("testdata/swagger.json")
	require.NoError(t, err)

	err = json.Unmarshal(input, &doc)
	require.NoError(t, err)

	output, err := json.Marshal(doc)
	require.NoError(t, err)

	require.JSONEq(t, string(input), string(output))
}
