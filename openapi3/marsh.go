package openapi3

import (
	"fmt"
	"net/url"
	"strings"

	goyaml "go.yaml.in/yaml/v3"

	"github.com/getkin/kin-openapi/internal/nodeindex"
)

func unmarshalError(jsonUnmarshalErr error) error {
	if before, after, found := strings.Cut(jsonUnmarshalErr.Error(), "Bis"); found && before != "" {
		before = strings.ReplaceAll(before, " Go struct ", " ")
		before = strings.ReplaceAll(before, " Go ", " ")
		return fmt.Errorf("%s%s", before, strings.ReplaceAll(after, "Bis", ""))
	}
	return jsonUnmarshalErr
}

// unmarshal decodes data into v, trying yaml first and json for what the yaml
// parser will not accept. It returns the document's parsed node tree when
// origins were asked for and the data took the yaml path (json input carries
// none), so the caller can retain it (see Loader.originTrees).
func unmarshal(data []byte, v any, includeOrigin bool, location *url.URL) (*originTree, error) {
	var file string
	if location != nil {
		file = location.String()
	}

	tree, yamlErr := unmarshalYAML(data, v, includeOrigin, file)
	if yamlErr == nil {
		return tree, nil
	}

	// Most importantly duplicate keys, which json resolves last-one-wins and
	// yaml rejects. Such documents load as they always did, without origins.
	// See https://github.com/getkin/kin-openapi/issues/680
	jsonErr := unmarshalJSON(data, v)
	if jsonErr == nil {
		return nil, nil
	}

	// If both unmarshaling attempts fail, return a new error that includes both errors
	return nil, fmt.Errorf("failed to unmarshal data: json error: %v, yaml error: %v", jsonErr, yamlErr)
}

// unmarshalYAML decodes data into v in one parse, straight into the types via
// UnmarshalYAML, with origins read off the nodes. A JSON document gets origins
// too, since JSON parses as YAML.
func unmarshalYAML(data []byte, v any, includeOrigin bool, file string) (*originTree, error) {
	var root goyaml.Node
	if err := goyaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	if err := prepareScalars(&root); err != nil {
		return nil, err
	}

	var ends *nodeindex.Index
	if includeOrigin {
		var release func()
		ends, release = installOriginState(&root, data, file)
		defer release()
	}

	if err := root.Decode(v); err != nil {
		return nil, err
	}
	if !includeOrigin {
		return nil, nil
	}

	// Retained so a $ref to an arbitrary top-level key can be decoded from its
	// own node; that path resolves through plain data, which carries no
	// positions.
	node := &root
	if root.Kind == goyaml.DocumentNode && len(root.Content) > 0 {
		node = root.Content[0]
	}
	stampRootOrigin(v, node)
	return &originTree{node: node, file: file, ends: ends}, nil
}

// installOriginState gives the decode about to run the file and index its
// UnmarshalYAML calls read, and returns the index along with the release to
// call when the decode ends.
//
// The state is the only thing a decode shares with another, so it is held under
// the lock, and a decode that produces no origins neither installs nor takes
// it. Releasing matters beyond the lock: the state holds an index over the
// whole node tree, which leaving it set would pin until the next decode
// replaced it, whether or not anything kept the tree. A caller that does need
// it gets it on the returned originTree.
func installOriginState(root *goyaml.Node, data []byte, file string) (*nodeindex.Index, func()) {
	originMu.Lock()
	// Ends are derived from the tree, the parser reporting only starts.
	ends := nodeindex.New(root, data)
	originState.Store(&originDecode{file: file, ends: ends})
	return ends, func() {
		originState.Store(nil)
		originMu.Unlock()
	}
}
