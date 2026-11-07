package openapi3

import (
	"errors"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v4"
)

// nameFields adds the path of the node each error is about to its message, so
// a construct error names the field and not only the line. The parser reports
// a position, which the tree turns back into a path.
//
// An error whose position no node starts at, or which is about the document
// itself, keeps the message it came with.
func nameFields(err error, root *yaml.Node) error {
	var loadErrs *yaml.LoadErrors
	if !errors.As(err, &loadErrs) {
		return err
	}

	for _, e := range loadErrs.Errors {
		if path, ok := nodePath(root, e.Mark.Line, e.Mark.Column); ok {
			e.Message += " in " + path
		}
	}
	return err
}

// nodePath returns the JSON pointer of the node starting at line and column,
// and false when no node starts there or the node is the document's root,
// which has no path to name.
func nodePath(root *yaml.Node, line, column int) (string, bool) {
	if root == nil || line == 0 || column == 0 {
		return "", false
	}

	node := root
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	path, ok := searchNode(node, line, column, nil)
	if !ok || path == "" {
		return "", false
	}
	return path, true
}

// searchNode returns the pointer of the node under node that starts at line and
// column, where path names node itself. The shallowest match wins: a mapping
// and its first key start in the same place, and it is the mapping that an
// error about the mapping's type is reported at.
func searchNode(node *yaml.Node, line, column int, path []string) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Line == line && node.Column == column {
		return pointerOf(path), true
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			if found, ok := searchNode(node.Content[i+1], line, column, append(path, node.Content[i].Value)); ok {
				return found, true
			}
		}
	case yaml.SequenceNode:
		for i, item := range node.Content {
			if found, ok := searchNode(item, line, column, append(path, strconv.Itoa(i))); ok {
				return found, true
			}
		}
	}
	return "", false
}

// pointerOf writes path as an RFC 6901 JSON pointer, the form a $ref names a
// component with.
func pointerOf(path []string) string {
	var b strings.Builder
	for _, name := range path {
		b.WriteByte('/')
		b.WriteString(strings.ReplaceAll(strings.ReplaceAll(name, "~", "~0"), "/", "~1"))
	}
	return b.String()
}
