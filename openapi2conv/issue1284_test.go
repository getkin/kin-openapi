package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

// Regression test for #1284.
//
// ToV3WithLoader builds doc3.Servers only inside `if host := doc2.Host;
// host != ""`. Swagger 2.0 makes `host` optional — "If the host is not
// included, the host serving the documentation is to be used (including
// the port)" — but `basePath` still applies, and OpenAPI 3 permits a
// relative server URL. So a v2 doc carrying `basePath` and no `host`
// converted to a v3 doc with no servers at all, and an absent `servers`
// means the default server URL of "/": /v1 was silently dropped.
//
// The reverse direction already handles the relative form: FromV3 reads
// parsedURL.Path off the first server as basePath whether or not that
// server has a host, so "/v1" round-trips V3 -> V2 but not V2 -> V3.
const issue1284HostlessSpec = `{
	"swagger": "2.0",
	"info": {"title": "Example", "version": "1"},
	"basePath": "/v1",
	"paths": {
		"/pets": {
			"get": {"responses": {"200": {"description": "OK"}}}
		}
	}
}`

func TestIssue1284_ToV3KeepsBasePathWithoutHost(t *testing.T) {
	t.Parallel()

	var doc2 openapi2.T
	require.NoError(t, json.Unmarshal([]byte(issue1284HostlessSpec), &doc2))

	// Pre-fix: len(doc3.Servers) == 0 here.
	doc3, err := openapi2conv.ToV3(&doc2)
	require.NoError(t, err)

	// No host, so the server URL must be the relative base path itself:
	// inventing a host would be wrong, dropping the path is worse.
	require.Len(t, doc3.Servers, 1, "basePath must survive as a server")
	require.Equal(t, "/v1", doc3.Servers[0].URL)
	require.Empty(t, doc3.Servers[0].Description)

	// A relative server URL has to survive validation, or the conversion
	// only relocates the failure to a later step.
	require.NoError(t, doc3.Validate(t.Context()))
}

// TestIssue1284_ToV3BasePathWithoutHostRoundTrips pins the asymmetry the
// report describes: FromV3 already takes the base path off a relative
// server, so converting back has to hand the basePath over again.
func TestIssue1284_ToV3BasePathWithoutHostRoundTrips(t *testing.T) {
	t.Parallel()

	var doc2 openapi2.T
	require.NoError(t, json.Unmarshal([]byte(issue1284HostlessSpec), &doc2))

	doc3, err := openapi2conv.ToV3(&doc2)
	require.NoError(t, err)

	back, err := openapi2conv.FromV3(doc3)
	require.NoError(t, err)

	// Pre-fix: back.BasePath == "" because doc3 carried no servers.
	require.Equal(t, "/v1", back.BasePath)
	require.Empty(t, back.Host, "no host may be invented in either direction")
}
