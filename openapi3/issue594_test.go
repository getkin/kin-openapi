package openapi3_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue594(t *testing.T) {
	uri, err := url.Parse("https://raw.githubusercontent.com/sendgrid/sendgrid-oai/c3aaa432b769faa47285166aca17c7ed2ea71787/oai_v3_stoplight.json")
	require.NoError(t, err)

	sl := openapi3.NewLoader()
	var doc *openapi3.T
	if false {
		doc, err = sl.LoadFromURI(uri)
	} else {
		doc, err = sl.LoadFromFile("testdata/oai_v3_stoplight.json")
	}
	require.NoError(t, err)

	doc.Info.Version = "1.2.3"
	doc.Paths.Value("/marketing/contacts/search/emails").Post = nil
	doc.Components.Schemas["full-segment"].Value.Example = nil

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}
