package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge_EmptySchema(t *testing.T) {
	schema := Schema{}
	flat := Merge(schema)
	require.Equal(t, &schema, flat)
}

func TestMerge_NotAllOf(t *testing.T) {
	schema := Schema{
		Title: "test",
	}
	flat := Merge(schema)
	require.Equal(t, &schema, flat)
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

	flat := Merge(schema)
	require.Equal(t, &schema, flat)
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

	flat := Merge(schema)
	require.Equal(t, &schema, flat)
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

	flat := Merge(schema)
	require.Len(t, flat.AllOf[0].Value.Properties, 2)
	require.Equal(t, obj1["description"], flat.AllOf[0].Value.Properties["description"])
	require.Equal(t, obj2["name"], flat.AllOf[0].Value.Properties["name"])
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

	flat := Merge(schema)
	require.Len(t, flat.AllOf[0].Value.Properties, 1)
	require.Equal(t, obj1["description"], flat.AllOf[0].Value.Properties["description"])
}
