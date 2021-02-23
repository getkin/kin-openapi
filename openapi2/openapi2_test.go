package openapi2

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadingSwagger(t *testing.T) {
	var swagger Swagger

	input, err := ioutil.ReadFile("testdata/swagger.json")
	require.NoError(t, err)

	err = json.Unmarshal(input, &swagger)
	require.NoError(t, err)

	output, err := json.Marshal(swagger)
	require.NoError(t, err)

	require.JSONEq(t, string(input), string(output))
}
