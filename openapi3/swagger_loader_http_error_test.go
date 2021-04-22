package openapi3

import (
	"net/url"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestLoadFromRemoteURLFailsWithHttpError(t *testing.T) {
	spec := []byte(`
{
    "openapi": "3.0.0",
    "info": {
        "title": "",
        "version": "1"
    },
    "paths": {
        "/test": {
            "post": {
                "responses": {
                    "default": {
                        "description": "test",
                        "headers": {
                            "X-TEST-HEADER": {
                                "$ref": "http://` + addr + `/components.openapi.json#/components/headers/CustomTestHeader"
                            }
                        }
                    }
                }
            }
        }
    }
}`)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://localhost:7965/components.openapi.json",
		httpmock.NewStringResponder(400, ""))
	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testdata/testfilename.openapi.json"})

	require.Nil(t, swagger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "request returned status code 400")
}
