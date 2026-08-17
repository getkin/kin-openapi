package openapi3filter

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func issue1112ObjectSchema(additional openapi3.AdditionalProperties) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:                 &openapi3.Types{"object"},
			Properties:           make(map[string]*openapi3.SchemaRef),
			AdditionalProperties: additional,
		},
	}
}

func issue1112Decode(t *testing.T, paramName, rawQuery string, schema *openapi3.SchemaRef) (any, bool, error) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "http://example.test/?"+rawQuery, nil)
	require.NoError(t, err)

	param := &openapi3.Parameter{
		Name:    paramName,
		In:      openapi3.ParameterInQuery,
		Style:   "deepObject",
		Explode: openapi3.Ptr(true),
		Schema:  schema,
	}

	return decodeStyledParameter(param, &RequestValidationInput{Request: req})
}

func TestIssue1112FreeFormDeepObject(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(
		t,
		"properties",
		"properties[vaccinated]=true&properties[color]=black&properties[coat_length]=large",
		schema,
	)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{
		"vaccinated":  "true",
		"color":       "black",
		"coat_length": "large",
	}, value)
	require.NoError(t, schema.Value.VisitJSON(value))
}

func TestIssue1112SingleFreeFormProperty(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(t, "properties", "properties[color]=black", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{"color": "black"}, value)
}

func TestIssue1112NestedFreeFormProperty(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(t, "properties", "properties[meta][color]=black", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{
		"meta": map[string]any{"color": "black"},
	}, value)
}

func TestIssue1112URLDecodedFreeFormValue(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(t, "properties", "properties[color]=blue%20green", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{"color": "blue green"}, value)
}

func TestIssue1112EmptyFreeFormValue(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(t, "properties", "properties[color]=", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{"color": ""}, value)
}

func TestIssue1112KnownPropertyKeepsTypedValue(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})
	schema.Value.Properties["limit"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
	}

	value, found, err := issue1112Decode(t, "properties", "properties[limit]=7&properties[color]=black", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{
		"limit": int64(7),
		"color": "black",
	}, value)
}

func TestIssue1112AdditionalPropertiesFalseDoesNotBecomeFreeForm(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(false)})

	value, found, err := issue1112Decode(t, "properties", "properties[color]=black", schema)

	require.NoError(t, err)
	require.False(t, found)
	require.Equal(t, map[string]any{}, value)
}

func TestIssue1112UnspecifiedAdditionalPropertiesRemainUnchanged(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{})

	value, found, err := issue1112Decode(t, "properties", "properties[color]=black", schema)

	require.NoError(t, err)
	require.False(t, found)
	require.Equal(t, map[string]any{}, value)
}

func TestIssue1112SchemaValuedAdditionalPropertiesStayTyped(t *testing.T) {
	integerSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
	}
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Schema: integerSchema})

	value, found, err := issue1112Decode(t, "properties", "properties[count]=7", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{"count": int64(7)}, value)
}

func TestIssue1112IgnoresUnrelatedQueryKeys(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(t, "properties", "other[color]=black", schema)

	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, value)
}

func TestIssue1112ParameterNameWithRegexpCharacters(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	value, found, err := issue1112Decode(t, "properties.v1", "properties.v1[color]=black", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{"color": "black"}, value)
}

func TestIssue1112RepeatedFreeFormValueStillRequiresIndexes(t *testing.T) {
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})

	_, found, err := issue1112Decode(t, "properties", "properties[tag]=one&properties[tag]=two", schema)

	require.False(t, found)
	require.Error(t, err)
	require.ErrorContains(t, err, "array items must be set with indexes")
}

func TestIssue1112NestedKnownFreeFormObject(t *testing.T) {
	metaSchema := issue1112ObjectSchema(openapi3.AdditionalProperties{Has: openapi3.Ptr(true)})
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{})
	schema.Value.Properties["meta"] = metaSchema

	value, found, err := issue1112Decode(t, "properties", "properties[meta][color]=black", schema)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, map[string]any{
		"meta": map[string]any{"color": "black"},
	}, value)
}

func TestIssue1112SchemaValuedNestedExtraneousPropertyRemainsIgnored(t *testing.T) {
	childSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: map[string]*openapi3.SchemaRef{
				"item1": {
					Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
				},
			},
		},
	}
	dynamicObject := issue1112ObjectSchema(openapi3.AdditionalProperties{Schema: childSchema})
	schema := issue1112ObjectSchema(openapi3.AdditionalProperties{})
	schema.Value.Properties["obj"] = dynamicObject

	value, found, err := issue1112Decode(t, "param", "param[obj][prop1][inexistent]=1", schema)

	require.NoError(t, err)
	require.False(t, found)
	require.Equal(t, map[string]any{
		"obj": map[string]any{
			"prop1": map[string]any{},
		},
	}, value)
}
