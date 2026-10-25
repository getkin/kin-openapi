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

// unmarshal decodes data into v. It returns the document's parsed node tree
// when includeOrigin is set and the data took the yaml path (json input carries
// no origins), so the caller can retain it (see Loader.originTrees).
func unmarshal(data []byte, v any, includeOrigin bool, location *url.URL) (*originTree, error) {
	var jsonErr, yamlErr error

	var file string
	if location != nil {
		file = location.String()
	}

	// One parse, straight into the types via UnmarshalYAML, with origins read
	// off the nodes. A JSON document gets origins too, since JSON parses as
	// YAML.
	// The state below is the only thing a decode shares with another, so a
	// decode that produces no origins takes no lock and runs beside them.
	//
	// It is released when the decode ends: the state exists to reach
	// UnmarshalYAML, so nothing needs it afterwards, and it holds an index over
	// the whole node tree, which leaving it set would pin until the next decode
	// replaced it, whether or not anything kept the tree. A caller that does
	// need it gets it on the returned originTree.
	if includeOrigin {
		originMu.Lock()
		defer originMu.Unlock()
		defer originState.Store(nil)
	}
	var ends *nodeindex.Index
	var root goyaml.Node
	if yamlErr = goyaml.Unmarshal(data, &root); yamlErr == nil {
		yamlErr = prepareScalars(&root)
	}
	if yamlErr == nil {
		if includeOrigin {
			// Ends are derived from the tree, the parser reporting only starts.
			ends = nodeindex.New(&root, data)
			originState.Store(&originDecode{file: file, ends: ends})
		}
		if yamlErr = root.Decode(v); yamlErr == nil {
			if !includeOrigin {
				return nil, nil
			}
			// Retained so a $ref to an arbitrary top-level key can be decoded
			// from its own node; that path resolves through plain data, which
			// carries no positions.
			if root.Kind == goyaml.DocumentNode && len(root.Content) > 0 {
				stampRootOrigin(v, root.Content[0])
				return &originTree{node: root.Content[0], file: file, ends: ends}, nil
			}
			stampRootOrigin(v, &root)
			return &originTree{node: &root, file: file, ends: ends}, nil
		}
	}

	// Fall back to the json path for what the yaml parser will not accept --
	// most importantly duplicate keys, which json resolves last-one-wins and
	// yaml rejects. Such documents load as they always did, without origins.
	// See https://github.com/getkin/kin-openapi/issues/680
	if jsonErr = unmarshalJSON(data, v); jsonErr == nil {
		return nil, nil
	}

	// If both unmarshaling attempts fail, return a new error that includes both errors
	return nil, fmt.Errorf("failed to unmarshal data: json error: %v, yaml error: %v", jsonErr, yamlErr)
}
