package openapi3conv

import (
	"fmt"
	"io"

	"github.com/getkin/kin-openapi/openapi3"
)

// DefaultTargetVersion is the OpenAPI version string written when bumping the
// document version. Set to the current 3.1 patch release; the OAI upgrade
// guide uses the same.
const DefaultTargetVersion = "3.1.1"

// UpgradeOptions controls per-pass behaviour.
type UpgradeOptions struct {
	// SkipVersionBump leaves doc.OpenAPI unchanged. Useful for consumers
	// that want representations canonicalized while preserving the stated
	// version (e.g., a pre-diff normalization step).
	SkipVersionBump bool

	// TargetVersion is the version string written when SkipVersionBump is
	// false. Defaults to DefaultTargetVersion.
	TargetVersion string

	// Verbose, if non-nil, receives one line per rewrite for debugging.
	Verbose io.Writer
}

// UpgradeTo31 rewrites every 3.0-form construct in doc into its 3.1 form
// in place: bumps the version, replaces nullable: true with type arrays,
// replaces boolean exclusiveMinimum/exclusiveMaximum with numeric form,
// replaces example with examples. Idempotent on already-3.1 documents.
//
// Returns an error only if the document is structurally invalid (e.g., nil
// document).
func UpgradeTo31(doc *openapi3.T) error {
	return UpgradeTo31WithOptions(doc, UpgradeOptions{})
}

// UpgradeTo31WithOptions is the variant with explicit options.
func UpgradeTo31WithOptions(doc *openapi3.T, opts UpgradeOptions) error {
	if doc == nil {
		return fmt.Errorf("openapi3conv: doc is nil")
	}

	target := opts.TargetVersion
	if target == "" {
		target = DefaultTargetVersion
	}

	w := &walker{
		visited: map[*openapi3.Schema]struct{}{},
		opts:    opts,
	}

	if !opts.SkipVersionBump && doc.OpenAPI != target {
		w.logf("openapi: %s -> %s", doc.OpenAPI, target)
		doc.OpenAPI = target
	}

	w.walkDoc(doc)
	return nil
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
	fmt.Fprintf(w.opts.Verbose, format+"\n", args...)
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

	if doc.Paths != nil {
		for _, pathItem := range doc.Paths.Map() {
			w.walkPathItem(pathItem)
		}
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
		alreadyHasNull := false
		for _, t := range *s.Type {
			if t == openapi3.TypeNull {
				alreadyHasNull = true
				break
			}
		}
		if !alreadyHasNull {
			newTypes := append(openapi3.Types(nil), *s.Type...)
			newTypes = append(newTypes, openapi3.TypeNull)
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
