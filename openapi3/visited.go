package openapi3

import "context"

func newVisited() *visitedComponent {
	return &visitedComponent{
		header: make(map[*Header]struct{}),
		schema: make(map[*Schema]struct{}),
	}
}

type visitedComponent struct {
	header map[*Header]struct{}
	schema map[*Schema]struct{}
}

type ctxKey struct{}

// withContext returns a copy of ctx with visitedComponent associated. If an instance
// of visitedComponent is already in the context, the context is not updated.
//
// Returned ctx can be passed to children function calls and used to check whether
// component already visited. For instance:
//
//	ctx := context.Background()
//	ctx = newVisited().withContext(ctx)
//	...
//	doc.deferSchemaRecursively(ctx, schema)
//	func (doc *T) deferSchemaRecursively(ctx context.Context, s *Schema) {
//	    if s == nil || isVisitedSchema(ctx, s) {
//	        return
//	    }
//	}
func (v *visitedComponent) withContext(ctx context.Context) context.Context {
	if visited, ok := ctx.Value(ctxKey{}).(*visitedComponent); ok {
		if visited == v {
			// Do not store the same object.
			return ctx
		}
	}

	return context.WithValue(ctx, ctxKey{}, v)
}

// visitedCtx returns the visitedComponent associated with the ctx. If no one
// is associated, a new visitedComponent is returned.
//
// The ctx should be initialized with method `withContext` first.
func visitedCtx(ctx context.Context) *visitedComponent {
	if v, ok := ctx.Value(ctxKey{}).(*visitedComponent); ok {
		return v
	}

	return newVisited()
}

// isVisitedHeader returns `true` if the *Header pointer was already visited
// otherwise it returns `false`
func isVisitedHeader(ctx context.Context, h *Header) bool {
	visited := visitedCtx(ctx)

	if _, ok := visited.header[h]; ok {
		return true
	}

	visited.header[h] = struct{}{}
	return false
}

// isVisitedHeader returns `true` if the *Schema pointer was already visited
// otherwise it returns `false`
func isVisitedSchema(ctx context.Context, s *Schema) bool {
	visited := visitedCtx(ctx)

	if _, ok := visited.schema[s]; ok {
		return true
	}

	visited.schema[s] = struct{}{}
	return false
}
