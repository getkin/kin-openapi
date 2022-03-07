package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathsValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr string
	}{
		{
			name: "ok, empty paths",
			spec: `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
paths:
  /pets:
`,
		},
		{
			name: "operation ids are not unique, same path",
			spec: `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
paths:
  /pets:
    post:
      operationId: createPet
      responses:
        201:
          description: "entity created"
    delete:
      operationId: createPet
      responses:
        204:
          description: "entity deleted"
`,
			wantErr: `operations "DELETE /pets" and "POST /pets" have the same operation id "createPet"`,
		},
		{
			name: "operation ids are not unique, different paths",
			spec: `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
paths:
  /pets:
    post:
      operationId: createPet
      responses:
        201:
          description: "entity created"
  /users:
    post:
      operationId: createPet
      responses:
        201:
          description: "entity created"
`,
			wantErr: `operations "POST /pets" and "POST /users" have the same operation id "createPet"`,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			doc, err := NewLoader().LoadFromData([]byte(tt.spec))
			require.NoError(t, err)

			err = doc.Paths.Validate(context.Background())
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Equal(t, tt.wantErr, err.Error())
		})
	}
}
