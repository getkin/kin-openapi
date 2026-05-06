package openapi3conv

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// DefaultTargetVersion is the OpenAPI version string written when bumping the
// document version. The OAI upgrade guide uses the current 3.1 patch release.
const DefaultTargetVersion = "3.1.1"

// UpgradeOptions controls per-pass behaviour.
type UpgradeOptions struct {
	// Target is the version string written into doc.OpenAPI after the
	// canonicalization pass. Defaults to DefaultTargetVersion. Currently any
	// 3.x version is accepted; representational rewrites only exist for
	// 3.0 → 3.1, since 3.1 → 3.2 is purely additive (no breaking changes).
	// Future minor versions that introduce representational changes can
	// extend the dispatch in Upgrade.
	Target string

	// Verbose, if non-nil, receives one line per rewrite for debugging.
	Verbose io.Writer
}

// Upgrade canonicalizes doc into the representation of opts.Target in place.
//
// The schema-level rewrites the walker applies (nullable → type array, boolean
// exclusive bounds → numeric, example → examples) are idempotent and
// convergent on the 3.1+ form. Calling Upgrade on an already-3.1 (or 3.2)
// document is a no-op aside from the version bump.
//
// Cross-major upgrades are not supported. OpenAPI 4 (or any future major
// version) will require a separate package mirroring the openapi2conv
// pattern (which converts Swagger 2.0 documents to OpenAPI 3.0). Returning
// an error here keeps that boundary explicit.
func Upgrade(doc *openapi3.T, opts UpgradeOptions) error {
	if doc == nil {
		return fmt.Errorf("openapi3conv: doc is nil")
	}

	target := opts.Target
	if target == "" {
		target = DefaultTargetVersion
	}

	srcMajor, srcMinor, err := parseVersion(doc.OpenAPI)
	if err != nil {
		return fmt.Errorf("openapi3conv: invalid doc.OpenAPI %q: %w", doc.OpenAPI, err)
	}
	tgtMajor, tgtMinor, err := parseVersion(target)
	if err != nil {
		return fmt.Errorf("openapi3conv: invalid Target %q: %w", target, err)
	}

	if srcMajor != tgtMajor {
		return fmt.Errorf(
			"openapi3conv: cross-major upgrade not supported (%s -> %s); "+
				"a separate package is the right home for cross-major conversions "+
				"(see openapi2conv for the existing 2 -> 3 pattern)",
			doc.OpenAPI, target,
		)
	}
	if tgtMinor < srcMinor {
		return fmt.Errorf("openapi3conv: cannot downgrade %s to %s", doc.OpenAPI, target)
	}

	w := &walker{
		visited: map[*openapi3.Schema]struct{}{},
		opts:    opts,
	}
	w.walkDoc(doc)

	if doc.OpenAPI != target {
		w.logf("openapi: %s -> %s", doc.OpenAPI, target)
		doc.OpenAPI = target
	}
	return nil
}

// UpgradeTo31 is a convenience wrapper for Upgrade with Target = "3.1.1".
// Idempotent on already-3.1 documents.
func UpgradeTo31(doc *openapi3.T) error {
	return Upgrade(doc, UpgradeOptions{})
}

// UpgradeSchema canonicalizes a single schema (and its descendants) in place.
// Exposed for callers that need to upgrade a sub-tree rather than a full
// document — e.g., a diff tool comparing isolated schemas.
func UpgradeSchema(s *openapi3.Schema) {
	if s == nil {
		return
	}
	w := &walker{visited: map[*openapi3.Schema]struct{}{}}
	w.walkSchema(s)
}

// parseVersion splits an OpenAPI version string ("3.0.3", "3.1.1", "3.2.0")
// into major and minor integers. Patch and pre-release suffixes are ignored.
func parseVersion(v string) (major, minor int, err error) {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("expected MAJOR.MINOR[.PATCH], got %q", v)
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid major %q: %w", parts[0], err)
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minor %q: %w", parts[1], err)
	}
	return major, minor, nil
}

// walker carries cycle-tracking state and verbose output across the schema
// graph. Each *Schema is visited at most once.
type walker struct {
	visited map[*openapi3.Schema]struct{}
	opts    UpgradeOptions
}

func (w *walker) logf(format string, args ...any) {
	if w.opts.Verbose == nil {
		return
	}
	fmt.Fprintf(w.opts.Verbose, format, args...)
	fmt.Fprintln(w.opts.Verbose)
}

// walkDoc visits every Schema reachable from the document root.
func (w *walker) walkDoc(doc *openapi3.T) {
	if doc.Components != nil {
		for _, sr := range doc.Components.Schemas {
			w.walkSchemaRef(sr)
		}
		for _, pr := range doc.Components.Parameters {
			if pr != nil && pr.Value != nil {
				w.walkSchemaRef(pr.Value.Schema)
				for _, mt := range pr.Value.Content {
					w.walkMediaType(mt)
				}
			}
		}
		for _, hr := range doc.Components.Headers {
			if hr != nil && hr.Value != nil {
				w.walkSchemaRef(hr.Value.Schema)
				for _, mt := range hr.Value.Content {
					w.walkMediaType(mt)
				}
			}
		}
		for _, rb := range doc.Components.RequestBodies {
			if rb != nil && rb.Value != nil {
				for _, mt := range rb.Value.Content {
					w.walkMediaType(mt)
				}
			}
		}
		for _, rr := range doc.Components.Responses {
			if rr != nil && rr.Value != nil {
				for _, mt := range rr.Value.Content {
					w.walkMediaType(mt)
				}
				for _, hr := range rr.Value.Headers {
					if hr != nil && hr.Value != nil {
						w.walkSchemaRef(hr.Value.Schema)
					}
				}
			}
		}
	}

	for _, pathItem := range doc.Paths.Map() {
		w.walkPathItem(pathItem)
	}

	for _, pathItem := range doc.Webhooks {
		w.walkPathItem(pathItem)
	}
}

func (w *walker) walkPathItem(pathItem *openapi3.PathItem) {
	if pathItem == nil {
		return
	}
	for _, pr := range pathItem.Parameters {
		if pr != nil && pr.Value != nil {
			w.walkSchemaRef(pr.Value.Schema)
		}
	}
	for _, op := range pathItem.Operations() {
		w.walkOperation(op)
	}
}

func (w *walker) walkOperation(op *openapi3.Operation) {
	if op == nil {
		return
	}
	for _, pr := range op.Parameters {
		if pr != nil && pr.Value != nil {
			w.walkSchemaRef(pr.Value.Schema)
			for _, mt := range pr.Value.Content {
				w.walkMediaType(mt)
			}
		}
	}
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		for _, mt := range op.RequestBody.Value.Content {
			w.walkMediaType(mt)
		}
	}
	if op.Responses != nil {
		for _, rr := range op.Responses.Map() {
			if rr == nil || rr.Value == nil {
				continue
			}
			for _, mt := range rr.Value.Content {
				w.walkMediaType(mt)
			}
			for _, hr := range rr.Value.Headers {
				if hr != nil && hr.Value != nil {
					w.walkSchemaRef(hr.Value.Schema)
				}
			}
		}
	}
	for _, cb := range op.Callbacks {
		if cb == nil || cb.Value == nil {
			continue
		}
		for _, pathItem := range cb.Value.Map() {
			w.walkPathItem(pathItem)
		}
	}
}

func (w *walker) walkMediaType(mt *openapi3.MediaType) {
	if mt == nil {
		return
	}
	w.walkSchemaRef(mt.Schema)
}

func (w *walker) walkSchemaRef(sr *openapi3.SchemaRef) {
	if sr == nil || sr.Value == nil {
		return
	}
	w.walkSchema(sr.Value)
}

func (w *walker) walkSchema(s *openapi3.Schema) {
	if s == nil {
		return
	}
	if _, seen := w.visited[s]; seen {
		return
	}
	w.visited[s] = struct{}{}

	// Apply the rewrites at this node before descending — order doesn't
	// matter, the transformations are independent.
	w.rewriteNullable(s)
	w.rewriteExclusiveBounds(s)
	w.rewriteExample(s)

	// Recurse into every child schema.
	for _, sub := range s.Properties {
		w.walkSchemaRef(sub)
	}
	w.walkSchemaRef(s.Items)
	if s.AdditionalProperties.Schema != nil {
		w.walkSchemaRef(s.AdditionalProperties.Schema)
	}
	for _, sub := range s.AllOf {
		w.walkSchemaRef(sub)
	}
	for _, sub := range s.OneOf {
		w.walkSchemaRef(sub)
	}
	for _, sub := range s.AnyOf {
		w.walkSchemaRef(sub)
	}
	w.walkSchemaRef(s.Not)
	for _, sub := range s.PatternProperties {
		w.walkSchemaRef(sub)
	}
}

// rewriteNullable converts `nullable: true` into a type-array form.
//
//   - With a non-empty Type: append "null" (deduped).
//   - Without a Type: drop nullable; the spec then accepts any type (no Type
//     restriction), which subsumes null. This matches the OAI guide's silence
//     on the "no type" edge case while keeping the rewrite lossless.
func (w *walker) rewriteNullable(s *openapi3.Schema) {
	if !s.Nullable {
		return
	}
	if s.Type != nil && len(*s.Type) > 0 {
		if !slices.Contains(*s.Type, openapi3.TypeNull) {
			newTypes := append(*s.Type, openapi3.TypeNull)
			s.Type = &newTypes
		}
	}
	w.logf("nullable: true -> dropped (Types=%v)", s.Type)
	s.Nullable = false
}

// rewriteExclusiveBounds converts the 3.0 boolean modifier into the 3.1
// numeric form:
//
//	minimum: x, exclusiveMinimum: true   ->  exclusiveMinimum: x  (Min cleared)
//	exclusiveMinimum: false              ->  field dropped (default)
//
// Mirror logic for maximum.
func (w *walker) rewriteExclusiveBounds(s *openapi3.Schema) {
	// Lower bound.
	if s.ExclusiveMin.Bool != nil {
		if *s.ExclusiveMin.Bool && s.Min != nil {
			v := *s.Min
			s.ExclusiveMin = openapi3.ExclusiveBound{Value: &v}
			s.Min = nil
			w.logf("exclusiveMinimum: true + minimum: %v -> exclusiveMinimum: %v (numeric)", v, v)
		} else {
			// false, or true without paired minimum — either way, the
			// boolean form has no meaning in 3.1. Drop it.
			s.ExclusiveMin = openapi3.ExclusiveBound{}
			w.logf("exclusiveMinimum: <bool> -> dropped")
		}
	}

	// Upper bound.
	if s.ExclusiveMax.Bool != nil {
		if *s.ExclusiveMax.Bool && s.Max != nil {
			v := *s.Max
			s.ExclusiveMax = openapi3.ExclusiveBound{Value: &v}
			s.Max = nil
			w.logf("exclusiveMaximum: true + maximum: %v -> exclusiveMaximum: %v (numeric)", v, v)
		} else {
			s.ExclusiveMax = openapi3.ExclusiveBound{}
			w.logf("exclusiveMaximum: <bool> -> dropped")
		}
	}
}

// rewriteExample converts the singular `example` field into the plural
// `examples` array on Schema Objects. The plural form is required in 3.1.
//
// Note: `example`/`examples` on Parameter and MediaType objects use a
// different shape (a map of named Example refs in 3.0+ that did not change
// for 3.1) and are not touched here.
func (w *walker) rewriteExample(s *openapi3.Schema) {
	if s.Example == nil {
		return
	}
	// Preserve the existing examples array; append to it. Some 3.0 specs
	// do set both example and examples (the latter as a vendor-flavored
	// extension); we keep both.
	s.Examples = append(s.Examples, s.Example)
	s.Example = nil
	w.logf("example: <v> -> examples: [..., <v>]")
}
