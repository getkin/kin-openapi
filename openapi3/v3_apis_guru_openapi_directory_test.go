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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3conv"
)

var goldens = filepath.Join("testdata", "apis_guru_openapi_directory")

func newUnderscorer() *strings.Replacer {
	var chars []string
	for c := range strings.SplitSeq("\\/_-( ).~", "") {
		chars = append(chars, c, "_")
	}
	return strings.NewReplacer(chars...)
}

func isOpenAPIVersion(t testing.TB, path, str string) bool {
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	r := bufio.NewScanner(file)
	for r.Scan() {
		require.NoError(t, r.Err())
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
	shortName = strings.TrimSuffix(shortName, "__validatebis")
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
	// encoding/json/v2 randomly picks "cannot" or "unable to" once per process.
	got = strings.ReplaceAll(got, "json: unable to ", "json: cannot ")
	if !strings.HasSuffix(got, "\n") {
		got += "\n"
	}

	if expectedStr != got {
		err := os.WriteFile(errf, []byte(got), 0644)
		require.NoError(t, err)

		require.Equal(t, expectedStr, got)
	}
}

// apisGuruDirectory returns the commit of APIs Guru's documents to use, and
// the directories they are downloaded to.
func apisGuruDirectory() (commit, dirName, targetDir string) {
	commit = os.Getenv("APISGURU_COMMIT")
	if commit == "" {
		commit = "f7207cf0a5c56081d275ebae4cf615249323385d" // On 2026-04-19
	}
	dirName = "APIs-guru-openapi-directory-" + commit[0:7]
	targetDir = filepath.Join("testdata", dirName)
	return
}

// apisGuruAPIs downloads APIs Guru's documents if needed and returns the
// directory holding them.
func apisGuruAPIs(tb testing.TB) string {
	if testing.Short() {
		tb.Skip("skipping APIs Guru's large sets of documents")
	}

	commit, dirName, targetDir := apisGuruDirectory()

	if _, err := os.Stat(targetDir); errors.Is(err, os.ErrNotExist) {
		req, err := http.NewRequestWithContext(tb.Context(), "GET", "https://github.com/APIs-guru/openapi-directory/tarball/"+commit, nil)
		require.NoError(tb, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(tb, err)
		defer resp.Body.Close()
		require.Equal(tb, http.StatusOK, resp.StatusCode)

		gzr, err := gzip.NewReader(resp.Body)
		require.NoError(tb, err)
		defer gzr.Close()
		tr := tar.NewReader(gzr)

		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			require.NoError(tb, err)

			target := filepath.Join(targetDir, header.Name)
			switch header.Typeflag {
			case tar.TypeDir:
				err := os.MkdirAll(target, 0755)
				require.NoError(tb, err)
			case tar.TypeReg:
				err := os.MkdirAll(filepath.Dir(target), 0755)
				require.NoError(tb, err)

				func() {
					f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
					require.NoError(tb, err)
					defer f.Close()

					_, err = io.Copy(f, tr)
					require.NoError(tb, err)
				}()
			}
		}
	}

	return filepath.Join(targetDir, dirName, "APIs")
}

func TestV3ApisGuruOpenapiDirectory(t *testing.T) {
	if !openapi3.UsingJSONv2 {
		t.Skip("skipping as goldens record errors from encoding/json/v2 (Go 1.27+)")
	}
	t.Parallel()

	root := apisGuruAPIs(t)

	underscorer := newUnderscorer()
	var checkedMu sync.Mutex
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

	t.Cleanup(func() {
		err := filepath.Walk(goldens, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				shortName := shortNameFromPath(path)
				delete(checked, shortName)
			}
			return nil
		})
		require.NoError(t, err)

		files, err := filepath.Glob(goldens + "/*")
		require.NoError(t, err)
		for _, file := range files {
			shortName := shortNameFromPath(file)
			if _, ok := checked[shortName]; ok || disabled(shortName) {
				err := os.Remove(file)
				require.NoError(t, err)
			}
		}
	})

	for _, path := range paths {
		shortName := underscorer.Replace(strings.TrimPrefix(path, root)[1:])

		t.Run(shortName, func(t *testing.T) {
			if disabled(shortName) {
				t.SkipNow()
				return
			}
			t.Parallel()

			if !isOpenAPIVersion(t, path, "openapi: 3") {
				t.SkipNow()
				return
			}

			checkedMu.Lock()
			checked[shortName] = struct{}{}
			checkedMu.Unlock()

			loader := openapi3.NewLoader()
			loader.Context = t.Context()

			doc, err := loader.LoadFromFile(path)
			golden(t, err, shortName, "load")
			if doc != nil {
				var opts []openapi3.ValidationOption
				err = doc.Validate(loader.Context, opts...)
				golden(t, err, shortName, "validate")

				if err == nil {
					openapi3conv.Upgrade(doc, openapi3conv.WithWriter(t.Output()))
					err = doc.Validate(loader.Context, opts...)
					golden(t, err, shortName, "validatebis")
				}
			}
		})
	}
}

func BenchmarkV3ApisGuruOpenapiDirectory(b *testing.B) {
	root := apisGuruAPIs(b)

	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && isOpenAPIVersion(b, path, "openapi: 3") {
			paths = append(paths, path)
		}
		return nil
	})
	require.NoError(b, err)

	b.ReportAllocs()
	for b.Loop() {
		for _, path := range paths {
			loader := openapi3.NewLoader()
			_, _ = loader.LoadFromFile(path)
		}
	}
	b.ReportMetric(float64(len(paths)), "documents")
}

func disabled(shortName string) bool {
	switch shortName {
	case "vvv keep these",
		"^^^ lines sorted":
		return true
	}
	return false
}
