//go:build go1.27 && goexperiment.jsonv2

package openapi2_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
)

func TestErrorsLocateValuesWithinMaps(t *testing.T) {
	var doc openapi2.T
	err := json.Unmarshal([]byte(`{"paths":{"/a":{"get":{"responses":{"200":{"schema":{"items":[1]}}}}}}}`), &doc)
	require.ErrorContains(t, err, `.paths.~1a.get.responses.200.schema.items`)
}
