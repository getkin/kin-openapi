// Package fuzzing holds what the fuzz targets share: finding the documents
// they seed their corpus with, and telling which documents survive YAML.
package fuzzing

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MaxSize keeps the seed corpus small enough that `go test ./...` still
// replays it on every run without noticeably slowing the suite down.
const MaxSize = 16 << 10

// Testdata returns every spec under the given testdata roots, however deep,
// so documents kept beside the tests that load them seed the fuzzers too.
// Fuzz corpora are skipped, and so is the downloaded APIs-guru directory,
// which APIsGuru reads.
func Testdata(tb testing.TB, roots ...string) [][]byte {
	tb.Helper()
	var seeds [][]byte
	for _, root := range roots {
		seeds = append(seeds, walk(tb, root, func(name string) bool {
			return name == "fuzz" || strings.HasPrefix(name, apisGuruPrefix)
		}, nil)...)
	}
	return seeds
}

const apisGuruPrefix = "APIs-guru-openapi-directory-"

// APIsGuru returns the documents of the APIs-guru corpus under testdata whose
// top-level keys include marker, e.g. "openapi:" or "swagger:". The corpus is
// downloaded by the TestV{2,3}ApisGuruOpenapiDirectory tests, and when it is
// absent this returns nothing. Under -short, which skips those tests, it
// returns nothing either, keeping the short suite short.
func APIsGuru(tb testing.TB, testdata, marker string) [][]byte {
	tb.Helper()
	if testing.Short() {
		return nil
	}
	dirs, err := filepath.Glob(filepath.Join(testdata, apisGuruPrefix+"*"))
	if err != nil {
		tb.Fatal(err)
	}
	var seeds [][]byte
	for _, dir := range dirs {
		// The tarball's own top-level directory carries the same name.
		root := filepath.Join(dir, filepath.Base(dir), "APIs")
		seeds = append(seeds, walk(tb, root, nil, func(data []byte) bool {
			return bytes.HasPrefix(data, []byte(marker)) || bytes.Contains(data, []byte("\n"+marker))
		})...)
	}
	return seeds
}

func walk(tb testing.TB, root string, skipDir func(string) bool, keep func([]byte) bool) [][]byte {
	tb.Helper()
	var seeds [][]byte
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDir != nil && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".yml", ".yaml", ".json":
		default:
			return nil
		}
		if info, err := d.Info(); err != nil || info.Size() > MaxSize {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if keep == nil || keep(data) {
			seeds = append(seeds, data)
		}
		return nil
	}); err != nil {
		tb.Fatal(err)
	}
	return seeds
}

// LosesInYAML reports whether v, decoded from JSON, holds something the YAML
// emitter does not write back faithfully, so a round trip through YAML text
// cannot be expected to preserve it:
//
//   - a "<<" key, which is written unquoted and so reads back as a merge;
//   - a string starting with a line break, which is written as a block
//     scalar that drops it.
func LosesInYAML(v any) bool {
	switch v := v.(type) {
	case string:
		return strings.HasPrefix(v, "\n") || strings.HasPrefix(v, "\r")
	case map[string]any:
		for k, e := range v {
			if k == "<<" || LosesInYAML(k) || LosesInYAML(e) {
				return true
			}
		}
	case []any:
		for _, e := range v {
			if LosesInYAML(e) {
				return true
			}
		}
	}
	return false
}
