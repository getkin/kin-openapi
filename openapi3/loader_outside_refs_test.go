package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOutsideRefs(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("testdata/303bis/service.yaml")
	require.NoError(t, err)
	require.NotNil(t, doc)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	require.Equal(t, "string", doc.
		Paths.Value("/service").
		Get.
		Responses.Value("200").Value.
		Content["application/json"].
		Schema.Value.
		Items.Value.
		AllOf[0].Value.
		Properties["created_at"].Value.
		Type)
}

func TestIssue423(t *testing.T) {
	spec := `
info:
  description: test
  title: test
  version: 0.0.0
openapi: 3.0.1
paths:
  /api/bundles:
    get:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Data'
components:
  schemas:
    Data:
      description: rbac Data
      properties:
        roles:
           $ref: https://raw.githubusercontent.com/kubernetes/kubernetes/132f29769dfecfc808adc58f756be43171054094/api/openapi-spec/swagger.json#/definitions/io.k8s.api.rbac.v1.RoleList
`

	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.LoadFromData([]byte(spec))
}
