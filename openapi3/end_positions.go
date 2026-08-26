package openapi3

import (
	"bytes"

	yaml "go.yaml.in/yaml/v3"
)

// endIndex answers where the block headed by a key ends.
//
// The parser reports where every node starts and nothing about where it stops,
// so the end is derived, in two passes over the tree.
//
// measure takes the largest line among a node's descendants. Reading that off
// the tree rather than from indentation is what makes a sequence item work: a
// parameter's key location is the item's first key, at the same column as the
// keys following it, so no column comparison can tell the end of the item from
// the start of its own second field.
//
// extend then grows that over lines belonging to the block that carry no node,
// which is the body of a block scalar. Together they agree with a parser's own
// end positions on 99.98% of blocks across a 4,240-file corpus; measure alone
// agrees on 99.60%, the difference being blocks that end in a `description: |`.
//
// Trailing blank and comment lines fall outside.
type endIndex struct {
	end map[*yaml.Node]int
	// anchorKey is the key heading each anchored node, so an alias can report
	// where the content it points at is defined.
	anchorKey map[*yaml.Node]*yaml.Node
	// lineLen is each line's length, giving the column a block's last line
	// stops at. indent is its leading-space count, or -1 when the line holds
	// nothing, which is what extend walks.
	lineLen []int
	indent  []int
}

// originEndsVar is the index for the decode in progress, alongside
// originFileVar. Ends are derived per file: a $ref into another file is decoded
// separately, against its own text.
var originEndsVar *endIndex

// newEndIndex indexes data's node tree. Returns nil when origins are off, in
// which case no end is ever asked for.
func newEndIndex(root *yaml.Node, data []byte) *endIndex {
	if !originEnabledVar || root == nil {
		return nil
	}
	ei := &endIndex{end: map[*yaml.Node]int{}, anchorKey: map[*yaml.Node]*yaml.Node{}}
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		line = bytes.TrimRight(line, "\r")
		ei.lineLen = append(ei.lineLen, len(line))
		if t := bytes.TrimRight(line, " \t"); len(t) == 0 {
			ei.indent = append(ei.indent, -1)
		} else {
			ei.indent = append(ei.indent, len(t)-len(bytes.TrimLeft(t, " ")))
		}
	}
	ei.measure(root)
	ei.extend(root)
	return ei
}

// extend grows each node's end over lines that belong to it but hang off no
// node of their own.
//
// measure can only see lines a descendant starts on, and the body of a block
// scalar has no nodes at all: `description: |` produces one scalar whose Line
// is where it starts, so a block ending in one is cut short. Since those lines
// are indented deeper than the block itself, indentation finds them.
//
// A blank line is included only when the block resumes below it, which is what
// separates a gap inside a block scalar from the gap after the block.
func (ei *endIndex) extend(n *yaml.Node) {
	for _, c := range n.Content {
		ei.extend(c)
	}
	e := ei.end[n]
	for {
		if ind := ei.indentOf(e); ind > n.Column-1 { // line e+1 is index e
			e++
			continue
		} else if ind == -1 {
			j := e
			for j < len(ei.indent) && ei.indentOf(j) == -1 {
				j++
			}
			if ei.indentOf(j) > n.Column-1 {
				e = j + 1
				continue
			}
		}
		break
	}
	ei.end[n] = e
}

// indentOf returns the indentation of the given 0-based line, or -1 when the
// line holds nothing or is past the end of the file.
func (ei *endIndex) indentOf(i int) int {
	if i < 0 || i >= len(ei.indent) {
		return -1
	}
	return ei.indent[i]
}

// measure records each node's last line, bottom up, and returns it.
func (ei *endIndex) measure(n *yaml.Node) int {
	if n == nil {
		return 0
	}
	last := n.Line
	if n.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(n.Content); i += 2 {
			if v := n.Content[i+1]; v.Anchor != "" {
				ei.anchorKey[v] = n.Content[i]
			}
		}
	}
	for _, c := range n.Content {
		if l := ei.measure(c); l > last {
			last = l
		}
	}
	ei.end[n] = last
	return last
}

// endOf returns the last line and column of the block node occupies.
func (ei *endIndex) endOf(node *yaml.Node) (int, int) {
	if ei == nil || node == nil {
		return 0, 0
	}
	line, ok := ei.end[node]
	if !ok || line < 1 || line > len(ei.lineLen) {
		return 0, 0
	}
	// The column just past the last character, matching how a parser reports
	// the position it stopped at.
	return line, ei.lineLen[line-1] + 1
}

// withEnd returns loc carrying the extent of the block node heads.
func withEnd(loc Location, node *yaml.Node) Location {
	loc.EndLine, loc.EndColumn = originEndsVar.endOf(node)
	return loc
}

// resolveAlias follows an alias to the node it points at, together with the key
// that node was defined under. An aliased schema is the anchored one, so its
// origin is where that content was written, not where it was referred to.
func resolveAlias(keyNode, valNode *yaml.Node) (*yaml.Node, *yaml.Node) {
	if valNode == nil || valNode.Kind != yaml.AliasNode || valNode.Alias == nil {
		return keyNode, valNode
	}
	if originEndsVar == nil {
		return keyNode, valNode.Alias
	}
	if k, ok := originEndsVar.anchorKey[valNode.Alias]; ok {
		return k, valNode.Alias
	}
	return keyNode, valNode.Alias
}
