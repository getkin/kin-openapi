package openapi3conv

import (
	"fmt"
	"io"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// latestTargetVersion is the OpenAPI version string written into doc.OpenAPI
// after canonicalization. Always the latest 3.x patch release the package
// knows about; bump when a new minor lands. The OAI guarantees strict
// compatibility for 3.x going forward (3.2.x, 3.3.x, ...), so a tool that
// handles 3.1 correctly handles later 3.x versions correctly too.
const latestTargetVersion = "3.2.0"

// Option configures an Upgrade pass. See WithWriter.
type Option func(*upgradeOptions)

// upgradeOptions is the internal carrier for Option functions. Kept private
// so the surface stays small and additive — new options are added by
// introducing a new WithX function.
type upgradeOptions struct {
	verbose io.Writer
}

// WithWriter routes one debug line per applied rewrite to w.
func WithWriter(w io.Writer) Option {
	return func(o *upgradeOptions) { o.verbose = w }
}

// Upgrade canonicalizes doc into the latest 3.x representation in place.
//
// The schema-level rewrites the walker applies (nullable → type array,
// boolean exclusive bounds → numeric, example → examples) are idempotent
// and convergent on the 3.1+ form. Calling Upgrade on an already-3.1 (or
// later) document is a no-op aside from the version string bump.
//
// Cross-major upgrades (3 → 4 if/when v4 ships) are not handled here; that
// belongs in a dedicated package mirroring the openapi2conv pattern.
//
// doc must be Validate()'d before calling Upgrade; passing an invalid
// document is undefined behaviour.
func Upgrade(doc *openapi3.T, opts ...Option) {
	if doc == nil {
		return
	}

	o := upgradeOptions{}
	for _, apply := range opts {
		apply(&o)
	}

	w := &walker{
		visited: map[*openapi3.Schema]struct{}{},
		opts:    o,
	}
	w.walkDoc(doc)

	if doc.OpenAPI != latestTargetVersion {
		w.logf("openapi: %s -> %s", doc.OpenAPI, latestTargetVersion)
		doc.OpenAPI = latestTargetVersion
	}
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
	opts    upgradeOptions
}

func (w *walker) logf(format string, args ...any) {
	if w.opts.verbose == nil {
		return
	}
	fmt.Fprintf(w.opts.verbose, format, args...)
	fmt.Fprintln(w.opts.verbose)
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
