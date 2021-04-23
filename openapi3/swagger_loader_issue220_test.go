package openapi3

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue220(t *testing.T) {
	// === RUN   TestIssue220
	//     swagger_loader_issue220_test.go:16: specPath: "testdata/my-openapi.json"
	//     swagger_loader_issue220_test.go:19: pwd: "D:\\a\\kin-openapi\\kin-openapi\\openapi3"
	//     swagger_loader_issue220_test.go:16: specPath: "testdata\\my-openapi.json"
	//     swagger_loader_issue220_test.go:19: pwd: "D:\\a\\kin-openapi\\kin-openapi\\openapi3"
	//     swagger_loader_issue220_test.go:24:
	//         	Error Trace:	swagger_loader_issue220_test.go:24
	//         	Error:      	Received unexpected error:
	//         	            	error resolving reference "my-other-openapi.json#/components/responses/DefaultResponse": err:"open my-other-openapi.json: The system cannot find the file specified.", path:"my-other-openapi.json", pwd:"D:\\a\\kin-openapi\\kin-openapi\\openapi3"
	//         	Test:       	TestIssue220
	// --- FAIL: TestIssue220 (0.00s)

	for _, specPath := range []string{
		"testdata/my-openapi.json",
		filepath.FromSlash("testdata/my-openapi.json"),
	} {
		t.Logf("specPath: %q", specPath)
		pwd, err := os.Getwd()
		require.NoError(t, err)
		t.Logf("pwd: %q", pwd)

		loader := NewSwaggerLoader()
		loader.IsExternalRefsAllowed = true
		doc, err := loader.LoadSwaggerFromFile(specPath)
		require.NoError(t, err)

		err = doc.Validate(loader.Context)
		require.NoError(t, err)

		require.Equal(t, "integer", doc.Paths["/foo"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Properties["bar"].Value.Type)
	}
}
