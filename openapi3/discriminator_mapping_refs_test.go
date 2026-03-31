package openapi3

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDiscriminatorMappingRefsInternalize tests that discriminator mapping refs
// are properly internalized when calling InternalizeRefs.
// This test demonstrates the issue that discriminator mapping values, which are
// JSON schema references serialized as plain strings, are not handled by InternalizeRefs.
func TestDiscriminatorMappingRefsInternalize(t *testing.T) {
	ctx := context.Background()

	// Load the spec with external discriminator mapping refs
	sl := NewLoader()
	sl.IsExternalRefsAllowed = true
	doc, err := sl.LoadFromFile("testdata/discriminator.yml")
	require.NoError(t, err, "loading test file")
	err = doc.Validate(ctx)
	require.NoError(t, err, "validating spec")

	// Verify the discriminator mapping refs are external before internalization
	schema := doc.Paths.Value("/").Get.Responses.Status(200).Value.Content.Get("application/json").Schema.Value.Items.Value
	require.NotNil(t, schema.Discriminator, "discriminator should exist")
	require.NotNil(t, schema.Discriminator.Mapping, "discriminator mapping should exist")

	// Before internalization, the mapping refs should be external
	fooMapping := schema.Discriminator.Mapping["foo"].Ref
	barMapping := schema.Discriminator.Mapping["bar"].Ref
	t.Logf("Before internalization - foo mapping: %s", fooMapping)
	t.Logf("Before internalization - bar mapping: %s", barMapping)

	require.True(t, strings.Contains(fooMapping, "ext.yml"), "foo mapping should reference external file before internalization")
	require.True(t, strings.Contains(barMapping, "ext.yml"), "bar mapping should reference external file before internalization")

	// Internalize the references
	doc.InternalizeRefs(ctx, nil)

	// Validate the internalized spec
	err = doc.Validate(ctx)
	require.NoError(t, err, "validating internalized spec")

	// After internalization, the mapping refs should be internal (#/components/...)
	schema = doc.Paths.Value("/").Get.Responses.Status(200).Value.Content.Get("application/json").Schema.Value.Items.Value
	fooMapping = schema.Discriminator.Mapping["foo"].Ref
	barMapping = schema.Discriminator.Mapping["bar"].Ref
	t.Logf("After internalization - foo mapping: %s", fooMapping)
	t.Logf("After internalization - bar mapping: %s", barMapping)

	// This is where the test fails currently - the mapping refs are NOT being internalized
	require.True(t, strings.HasPrefix(fooMapping, "#/components/"), "foo mapping should be internalized to #/components/...")
	require.True(t, strings.HasPrefix(barMapping, "#/components/"), "bar mapping should be internalized to #/components/...")
}
