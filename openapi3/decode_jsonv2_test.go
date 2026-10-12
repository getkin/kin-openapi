//go:build go1.27 && goexperiment.jsonv2

package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestDuplicateNamesAreRejected(t *testing.T) {
	var info openapi3.Info
	err := json.Unmarshal([]byte(`{"title":"a","title":"b","version":"1"}`), &info)
	require.ErrorContains(t, err, `duplicate object member name`)

	var doc openapi3.T
	err = json.Unmarshal([]byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"},"paths":{"/a":{},"/a":{}}}`), &doc)
	require.ErrorContains(t, err, `duplicate object member name`)
}

func TestErrorsLocateValuesWithinMaps(t *testing.T) {
	var doc openapi3.T
	err := json.Unmarshal([]byte(`{"paths":{"/a":{"get":{"responses":{"200":{"description":1}}}}}}`), &doc)
	require.ErrorContains(t, err, `.paths.~1a.get.responses.200.description`)
}
