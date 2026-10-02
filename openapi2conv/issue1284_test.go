package openapi2conv_test

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue1284(t *testing.T) {
	for _, tc := range []struct {
		name              string
		host              string
		basePath          string
		schemes           []string
		servers           openapi3.Servers
		roundTripBasePath string
		requestURI        string
	}{
		{name: "no host or base path"},
		{name: "schemes without host", schemes: []string{"http", "https"}},
		{name: "base path", basePath: "/v1", servers: openapi3.Servers{{URL: "/v1"}}},
		{name: "root", basePath: "/", servers: openapi3.Servers{{URL: "/"}}},
		{name: "trailing slash", basePath: "/v1/", servers: openapi3.Servers{{URL: "/v1/"}}},
		{name: "unicode", basePath: "/日本語", servers: openapi3.Servers{{URL: "/%E6%97%A5%E6%9C%AC%E8%AA%9E"}}},
		{name: "percent", basePath: "/v1%2Fpets", servers: openapi3.Servers{{URL: "/v1%252Fpets"}}},
		{name: "leading double slash", basePath: "//api.example.com/v1", servers: openapi3.Servers{{URL: "/.//api.example.com/v1"}}, roundTripBasePath: "/.//api.example.com/v1", requestURI: "//api.example.com/v1"},
		{name: "leading triple slash", basePath: "///v1", servers: openapi3.Servers{{URL: "/.///v1"}}, roundTripBasePath: "/.///v1", requestURI: "///v1"},
		{name: "query character", basePath: "/v1?pets", servers: openapi3.Servers{{URL: "/v1%3Fpets"}}},
		{name: "fragment character", basePath: "/v1#pets", servers: openapi3.Servers{{URL: "/v1%23pets"}}},
		{name: "scheme-like path", basePath: "/https://api.example.com/v1", servers: openapi3.Servers{{URL: "/https://api.example.com/v1"}}},
		{name: "base path with schemes", basePath: "/v1", schemes: []string{"http", "https"}, servers: openapi3.Servers{{URL: "/v1"}}},
		{name: "host", host: "api.example.com", basePath: "/v1", servers: openapi3.Servers{{URL: "https://api.example.com/v1"}}},
		{name: "host with schemes", host: "api.example.com", basePath: "/v1", schemes: []string{"http", "https"}, servers: openapi3.Servers{{URL: "http://api.example.com/v1"}, {URL: "https://api.example.com/v1"}}},
		{name: "host with percent", host: "api.example.com", basePath: "/v1%2Fpets", servers: openapi3.Servers{{URL: "https://api.example.com/v1%252Fpets"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var doc2 openapi2.T
			err := json.Unmarshal([]byte(`{"swagger":"2.0","info":{"title":"Example","version":"1"},"paths":{"/pets":{"get":{"responses":{"200":{"description":"OK"}}}}}}`), &doc2)
			require.NoError(t, err)
			doc2.Host, doc2.BasePath, doc2.Schemes = tc.host, tc.basePath, tc.schemes

			doc3, err := openapi2conv.ToV3(&doc2)
			require.NoError(t, err)
			require.Equal(t, tc.servers, doc3.Servers)
			require.NoError(t, doc3.Validate(t.Context()))

			roundTrip, err := openapi2conv.FromV3(doc3)
			require.NoError(t, err)
			require.Equal(t, tc.host, roundTrip.Host)
			expectedBasePath := tc.basePath
			if tc.roundTripBasePath != "" {
				expectedBasePath = tc.roundTripBasePath
			}
			require.Equal(t, expectedBasePath, roundTrip.BasePath)

			if tc.host == "" && len(doc3.Servers) != 0 {
				again, err := openapi2conv.ToV3(roundTrip)
				require.NoError(t, err)
				location := &url.URL{Scheme: "https", Host: "docs.example.com", Path: "/spec/swagger.json"}
				requestURI := tc.servers[0].URL
				if tc.requestURI != "" {
					requestURI = tc.requestURI
				}
				for _, servers := range []openapi3.Servers{doc3.Servers, again.Servers} {
					require.Len(t, servers, 1)
					server, err := url.Parse(servers[0].URL)
					require.NoError(t, err)
					require.Empty(t, server.Scheme)
					require.Empty(t, server.Host)
					resolved := location.ResolveReference(server)
					require.Equal(t, location.Host, resolved.Host)
					require.Equal(t, requestURI, resolved.RequestURI())
				}
			}
		})
	}
}
