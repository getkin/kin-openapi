package openapi3

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	goyaml "go.yaml.in/yaml/v3"
)

func unmarshalError(jsonUnmarshalErr error) error {
	if before, after, found := strings.Cut(jsonUnmarshalErr.Error(), "Bis"); found && before != "" && after != "" {
		before = strings.ReplaceAll(before, " Go struct ", " ")
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
	originMu.Lock()
	defer originMu.Unlock()
	// Released when the decode ends. They exist to carry state into
	// UnmarshalYAML, so nothing needs them afterwards, and originEndsVar holds
	// an index over the whole node tree: leaving it set pins that tree until
	// the next decode replaces it, whether or not anything kept the tree.
	// A caller that does need it gets it on the returned originTree.
	defer func() { originFileVar, originEnabledVar, originEndsVar = "", false, nil }()
	originFileVar, originEnabledVar = file, includeOrigin
	var root goyaml.Node
	if err := goyaml.Unmarshal(data, &root); err == nil {
		stripTimestamps(&root)
		// Ends are derived from the tree, the parser reporting only starts.
		originEndsVar = newEndIndex(&root, data)
		if err = root.Decode(v); err == nil {
			if !includeOrigin {
				return nil, nil
			}
			// Retained so a $ref to an arbitrary top-level key can be decoded
			// from its own node; that path resolves through plain data, which
			// carries no positions.
			if root.Kind == goyaml.DocumentNode && len(root.Content) > 0 {
				stampRootOrigin(v, root.Content[0])
				return &originTree{node: root.Content[0], file: file, ends: originEndsVar}, nil
			}
			stampRootOrigin(v, &root)
			return &originTree{node: &root, file: file, ends: originEndsVar}, nil
		}
		yamlErr = err
	} else {
		yamlErr = err
	}

	// Fall back to the json path for what the yaml parser will not accept --
	// most importantly duplicate keys, which json resolves last-one-wins and
	// yaml rejects. Such documents load as they always did, without origins.
	// See https://github.com/getkin/kin-openapi/issues/680
	if jsonErr = json.Unmarshal(data, v); jsonErr == nil {
		return nil, nil
	}

	// If both unmarshaling attempts fail, return a new error that includes both errors
	return nil, fmt.Errorf("failed to unmarshal data: json error: %v, yaml error: %v", jsonErr, yamlErr)
}
