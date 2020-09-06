package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExampleJSON(t *testing.T) {
	t.Log("Marshal *openapi3.Example to JSON")
	data, err := json.Marshal(example())
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Unmarshal *openapi3.Example from JSON")
	docA := &Example{}
	err = json.Unmarshal(exampleJSON, &docA)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Log("Ensure representations match")
	dataA, err := json.Marshal(docA)
	require.NoError(t, err)
	require.JSONEq(t, string(data), string(exampleJSON))
	require.JSONEq(t, string(data), string(dataA))
}

var exampleJSON = []byte(`
{
   "summary": "An example of a cat",
   "value": {
      "name": "Fluffy",
      "petType": "Cat",
      "color": "White",
      "gender": "male",
      "breed": "Persian"
   }
}
`)

func example() *Example {
	value := map[string]string{
		"name":    "Fluffy",
		"petType": "Cat",
		"color":   "White",
		"gender":  "male",
		"breed":   "Persian",
	}
	return &Example{
		Summary: "An example of a cat",
		Value:   value,
	}
}
