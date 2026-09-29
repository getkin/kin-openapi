package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
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
		{
			name: "ok, path parameter name matches template",
			spec: `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Repro
paths:
  /pets/{petId}:
    get:
      parameters:
        - name: petId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: ok
`,
		},
		{
			name: "path parameter name swapped with template placeholder",
			spec: `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Repro
paths:
  /pets/{petId}:
    get:
      operationId: showPetById
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: ok
`,
			wantErr: `operation GET /pets/{petId} must define exactly all path parameters (missing: [id petId])`,
		},
		{
			name: "leftover template placeholder after a matching name",
			spec: `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Repro
paths:
  /{a}/{b}:
    get:
      parameters:
        - name: a
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: ok
`,
			wantErr: `operation GET /{a}/{b} must define exactly all path parameters (missing: [b])`,
		},
		{
			name: "ok, gorilla mux regex decoration on placeholder",
			spec: `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Repro
paths:
  /params/{z:.*}:
    get:
      parameters:
        - name: z
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: ok
`,
		},
		{
			name: "ok, legacy trailing-wildcard decoration on placeholder",
			spec: `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Repro
paths:
  /files/{path*}:
    get:
      parameters:
        - name: path
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: ok
`,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			loader := openapi3.NewLoader()

			doc, err := loader.LoadFromData([]byte(tt.spec[1:]))
			require.NoError(t, err)

			err = doc.Paths.Validate(t.Context())
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr)
		})
	}
}
