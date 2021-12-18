package openapi2conv

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIssue440(t *testing.T) {
	doc2file, err := os.Open("testdata/swagger.json")
	require.NoError(t, err)
	defer doc2file.Close()
	var doc2 openapi2.T
	err = json.NewDecoder(doc2file).Decode(&doc2)
	require.NoError(t, err)

	doc3, err := ToV3(&doc2)
	require.NoError(t, err)
	err = doc3.Validate(context.Background())
	require.NoError(t, err)
	require.Equal(t, openapi3.Servers{
		{URL: "https://petstore.swagger.io/v2"},
		{URL: "http://petstore.swagger.io/v2"},
	}, doc3.Servers)

	doc2.Host = "your-bot-domain.de"
	doc2.Schemes = nil
	doc2.BasePath = ""
	doc3, err = ToV3(&doc2)
	require.NoError(t, err)
	err = doc3.Validate(context.Background())
	require.NoError(t, err)
	require.Equal(t, openapi3.Servers{
		{URL: "https://your-bot-domain.de/"},
	}, doc3.Servers)

	doc2.Host = "https://your-bot-domain.de"
	doc2.Schemes = nil
	doc2.BasePath = ""
	doc3, err = ToV3(&doc2)
	require.Error(t, err)
	require.Contains(t, err.Error(), `invalid host`)
}
