package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge_EmptySchema(t *testing.T) {
	schema := Schema{}
	merged := Merge(schema)
	require.Equal(t, &schema, merged)
}

func TestMerge_NotAllOf(t *testing.T) {
	schema := Schema{
		Title: "test",
	}
	merged := Merge(schema)
	require.Equal(t, &schema, merged)
}

func TestMerge_OneObjectNoProps(t *testing.T) {

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: Schemas{},
				},
			},
		},
	}

	merged := Merge(schema)
	require.Equal(t, &schema, merged)
}

func TestMerge_OneObjectOneProp(t *testing.T) {

	object := Schemas{}
	object["description"] = &SchemaRef{
		Value: &Schema{
			Type: "string",
		},
	}

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: object,
				},
			},
		},
	}

	merged := Merge(schema)
	require.Equal(t, &schema, merged)
}

func TestMerge_TwoObjects(t *testing.T) {

	obj1 := Schemas{}
	obj1["description"] = &SchemaRef{
		Value: &Schema{
			Type: "string",
		},
	}

	obj2 := Schemas{}
	obj2["name"] = &SchemaRef{
		Value: &Schema{
			Type: "string",
		},
	}

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: obj1,
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: obj2,
				},
			},
		},
	}

	merged := Merge(schema)
	require.Len(t, merged.AllOf, 1)
	require.Len(t, merged.AllOf[0].Value.Properties, 2)
	require.Equal(t, obj1["description"], merged.AllOf[0].Value.Properties["description"])
	require.Equal(t, obj2["name"], merged.AllOf[0].Value.Properties["name"])
}

func TestMerge_OverlappingProps(t *testing.T) {

	obj1 := Schemas{}
	obj1["description"] = &SchemaRef{
		Value: &Schema{
			Type: "string",
		},
	}

	obj2 := Schemas{}
	obj2["description"] = &SchemaRef{
		Value: &Schema{
			Type: "int",
		},
	}

	schema := Schema{
		AllOf: SchemaRefs{
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: obj1,
				},
			},
			&SchemaRef{
				Value: &Schema{
					Type:       "object",
					Properties: obj2,
				},
			},
		},
	}

	merged := Merge(schema)
	require.Len(t, merged.AllOf, 1)
	require.Len(t, merged.AllOf[0].Value.Properties, 1)
	require.Equal(t, obj1["description"], merged.AllOf[0].Value.Properties["description"])
}

func TestMerge_Required(t *testing.T) {

	ctx := context.Background()

	tests := []struct {
		filename string
	}{
		{"testdata/mergeSchemas/properties.yml"},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			// Load in the reference spec from the testdata
			sl := NewLoader()
			sl.IsExternalRefsAllowed = true
			doc, err := sl.LoadFromFile(test.filename)
			require.NoError(t, err, "loading test file")
			err = doc.Validate(ctx)
			require.NoError(t, err, "validating spec")

			merged := Merge(*doc.Paths["/products"].Get.Responses["200"].Value.Content["application/json"].Schema.Value)
			require.Len(t, merged.AllOf, 1)

			props := merged.AllOf[0].Value.Properties
			require.Len(t, props, 3)
			require.Contains(t, props, "id")
			require.Contains(t, props, "createdAt")
			require.Contains(t, props, "otherId")

			required := merged.AllOf[0].Value.Required
			require.Len(t, required, 2)
			require.Contains(t, required, "id")
			require.Contains(t, required, "otherId")
		})
	}
}
