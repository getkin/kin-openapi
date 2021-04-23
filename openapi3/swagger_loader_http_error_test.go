package openapi3

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFromRemoteURLFailsWithHttpError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "")
	}))
	defer ts.Close()

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
                                "$ref": "` + ts.URL + `/components.openapi.json#/components/headers/CustomTestHeader"
                            }
                        }
                    }
                }
            }
        }
    }
}`)

	loader := NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadSwaggerFromDataWithPath(spec, &url.URL{Path: "testdata/testfilename.openapi.json"})

	require.Nil(t, swagger)
	require.EqualError(t, err, fmt.Sprintf("error resolving reference \"%s/components.openapi.json#/components/headers/CustomTestHeader\": request returned status code 400", ts.URL))

	swagger, err = loader.LoadSwaggerFromData(spec)
	require.Nil(t, swagger)
	require.EqualError(t, err, fmt.Sprintf("error resolving reference \"%s/components.openapi.json#/components/headers/CustomTestHeader\": request returned status code 400", ts.URL))
}
