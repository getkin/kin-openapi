package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

// A Swagger document can write null wherever an object belongs, and every
// collection that accepts one reached a dereference. Converting such a
// document reports it rather than panicking: the member stays nil and the
// resulting document says so when validated.
func TestToV3NullMembers(t *testing.T) {
	t.Parallel()

	for _, doc2 := range []string{
		`{"paths":{"0":null}}`,
		`{"paths":{"0":{"get":{"responses":{"":null}}}}}`,
		`{"paths":{"0":{"get":{"responses":{"200":{"description":"d","headers":{"h":null}}}}}}}`,
		`{"paths":{"0":{"get":{"parameters":[null]}}}}`,
		`{"paths":{"0":{"parameters":[null],"get":{"responses":{"200":{"description":"d"}}}}}}`,
		`{"responses":{"x":null}}`,
		`{"parameters":{"x":null}}`,
		`{"definitions":{"x":null}}`,
		`{"securityDefinitions":{"x":null}}`,
	} {
		t.Run(doc2, func(t *testing.T) {
			t.Parallel()

			var doc openapi2.T
			require.NoError(t, json.Unmarshal([]byte(doc2), &doc))

			// Whatever comes back, nothing dereferences a null on the way.
			doc3, err := openapi2conv.ToV3(&doc)
			if err != nil {
				return
			}
			_, err = doc3.MarshalJSON()
			require.NoError(t, err)
		})
	}
}

// Converting a document leaves it as it was: a v3 header drops the in and name
// a v2 header carries, and the copy being converted is where that happens.
func TestToV3HeadersLeavesInputAlone(t *testing.T) {
	t.Parallel()

	const spec = `{"paths":{"/a":{"get":{"responses":{"200":{"description":"d","headers":{"h":{"type":"string"}}}}}}}}`

	var doc openapi2.T
	require.NoError(t, json.Unmarshal([]byte(spec), &doc))
	header := doc.Paths["/a"].Get.Responses.Value("200").Headers["h"]
	header.Name, header.In = "h", "header"

	_, err := openapi2conv.ToV3(&doc)
	require.NoError(t, err)
	require.Equal(t, "h", header.Name)
	require.Equal(t, "header", header.In)
}
