package openapi3_test

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

var goldens = filepath.Join("testdata", "apis_guru_openapi_directory")

func newUnderscorer() *strings.Replacer {
	var chars []string
	for c := range strings.SplitSeq("\\/_-( ).~", "") {
		chars = append(chars, c, "_")
	}
	return strings.NewReplacer(chars...)
}

func isOpenAPIVersion(t *testing.T, path, str string) bool {
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	r := bufio.NewScanner(file)
	for r.Scan() {
		if strings.Contains(r.Text(), str) {
			return true
		}
	}
	return false
}

func shortNameFromPath(path string) string {
	shortName := filepath.Base(path)
	shortName = strings.TrimSuffix(shortName, "__load")
	shortName = strings.TrimSuffix(shortName, "__validate")
	return shortName
}

func golden(t *testing.T, e error, shortName, task string) {
	errf := filepath.Join(goldens, shortName+"__"+task)

	if e == nil {
		_ = os.Remove(errf)
		return
	}

	expected, _ := os.ReadFile(errf)
	expectedStr := strings.ReplaceAll(string(expected), "\r\n", "\n")
	got := strings.ReplaceAll(e.Error(), "\r\n", "\n")
	if !strings.HasSuffix(got, "\n") {
		got += "\n"
	}

	if expectedStr != got {
		err := os.WriteFile(errf, []byte(got), 0644)
		require.NoError(t, err)

		require.Equal(t, expectedStr, got)
	}
}

func TestV3ApisGuruOpenapiDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping APIs Guru's large sets of documents")
	}

	commit := os.Getenv("APISGURU_COMMIT")
	if commit == "" {
		commit = "f7207cf0a5c56081d275ebae4cf615249323385d" // On 2026-04-19
	}
	dirName := "APIs-guru-openapi-directory-" + commit[0:7]
	targetDir := filepath.Join("testdata", dirName)

	if _, err := os.Stat(targetDir); errors.Is(err, os.ErrNotExist) {
		req, err := http.NewRequestWithContext(t.Context(), "GET", "https://github.com/APIs-guru/openapi-directory/tarball/"+commit, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		gzr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)
		defer gzr.Close()
		tr := tar.NewReader(gzr)

		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)

			target := filepath.Join(targetDir, header.Name)
			switch header.Typeflag {
			case tar.TypeDir:
				err := os.MkdirAll(target, 0755)
				require.NoError(t, err)
			case tar.TypeReg:
				err := os.MkdirAll(filepath.Dir(target), 0755)
				require.NoError(t, err)

				func() {
					f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
					require.NoError(t, err)
					defer f.Close()

					_, err = io.Copy(f, tr)
					require.NoError(t, err)
				}()
			}
		}
	}

	root := filepath.Join(targetDir, dirName, "APIs")

	underscorer := newUnderscorer()
	checked := make(map[string]struct{})

	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	require.NoError(t, err)
	t.Logf("found %v files in %q", len(paths), root)

	for _, path := range paths {
		shortName := underscorer.Replace(strings.TrimPrefix(path, root)[1:])

		if isOpenAPIVersion(t, path, "openapi: 3") {
			t.Run(shortName, func(t *testing.T) {
				if disabled(shortName) {
					t.SkipNow()
					return
				}
				checked[shortName] = struct{}{}
				t.Parallel()

				loader := openapi3.NewLoader()
				loader.Context = t.Context()

				doc, err := loader.LoadFromFile(path)
				golden(t, err, shortName, "load")
				if doc != nil {
					var opts []openapi3.ValidationOption
					err = doc.Validate(loader.Context, opts...)
					golden(t, err, shortName, "validate")
				}
			})
		}
	}

	err = filepath.Walk(goldens, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			shortName := shortNameFromPath(path)
			delete(checked, shortName)
		}
		return nil
	})
	require.NoError(t, err)

	files, err := filepath.Glob(goldens + "*")
	require.NoError(t, err)
	for _, file := range files {
		shortName := shortNameFromPath(file)
		if _, ok := checked[shortName]; ok || disabled(shortName) {
			err := os.Remove(file)
			require.NoError(t, err)
		}
	}

}

func disabled(shortName string) bool {
	switch shortName {
	case "vvv keep these",
		"flat_io_2_13_0_openapi_yaml",                                 // TODO: flaky
		"microsoft_com_cognitiveservices_Prediction_3_0_openapi_yaml", // TODO: flaky
		"microsoft_com_cognitiveservices_Training_3_0_openapi_yaml",   // TODO: flaky
		"microsoft_com_cognitiveservices_Training_3_1_openapi_yaml",   // TODO: flaky
		"microsoft_com_cognitiveservices_Training_3_2_openapi_yaml",   // TODO: flaky
		"mist_com_0_37_7_openapi_yaml",                                // TODO: flaky
		"ndhm_gov_in_ndhm_cm_0_5_openapi_yaml",                        // TODO: flaky
		"nexmo_com_voice_1_3_10_openapi_yaml",                         // TODO: flaky
		"nordigen_com_2_0__v2__openapi_yaml",                          // TODO: YAML dates in map keys https://github.com/invopop/yaml/issues/10
		"optimade_local_1_1_0_develop_openapi_yaml",                   // TODO: flaky
		"unicourt_com_1_0_0_openapi_yaml",                             // TODO: YAML dates in map keys https://github.com/invopop/yaml/issues/10
		"zuora_com_2021_08_20_openapi_yaml",                           // TODO: YAML dates in map keys https://github.com/invopop/yaml/issues/10
		"^^^ lines sorted":
		return true
	}
	return false
}
