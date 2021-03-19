// Package gorillamux implements a router.
//
// It differs from the legacy router:
// * it provides somewhat granular errors: "path not found", "method not allowed".
// * it handles matching routes with extensions (e.g. /books/{id}.json)
// * it handles path patterns with a different syntax (e.g. /params/{x}/{y}/{z:.*})
package gorillamux

import (
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/gorilla/mux"
)

// Router helps link http.Request.s and an OpenAPIv3 spec
type Router struct {
	muxes  []*mux.Route
	routes []*routers.Route
}

// NewRouter creates a gorilla/mux router.
// Assumes spec is .Validate()d
// TODO: Handle/HandlerFunc + ServeHTTP (When there is a match, the route variables can be retrieved calling mux.Vars(request))
func NewRouter(doc *openapi3.Swagger) (routers.Router, error) {
	type srv struct {
		scheme, host, base string
		server             *openapi3.Server
	}
	servers := make([]srv, 0, len(doc.Servers))
	for _, server := range doc.Servers {
		u, err := url.Parse(bEncode(server.URL))
		if err != nil {
			return nil, err
		}
		path := bDecode(u.EscapedPath())
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
		servers = append(servers, srv{
			host:   bDecode(u.Host), //u.Hostname()?
			base:   path,
			scheme: bDecode(u.Scheme),
			server: server,
		})
	}
	if len(servers) == 0 {
		servers = append(servers, srv{})
	}
	muxRouter := mux.NewRouter() /*.UseEncodedPath()?*/
	r := &Router{}
	for _, path := range orderedPaths(doc.Paths) {
		pathItem := doc.Paths[path]

		operations := pathItem.Operations()
		methods := make([]string, 0, len(operations))
		for method := range operations {
			methods = append(methods, method)
		}
		sort.Strings(methods)

		for _, s := range servers {
			muxRoute := muxRouter.Path(s.base + path).Methods(methods...)
			if scheme := s.scheme; scheme != "" {
				muxRoute.Schemes(scheme)
			}
			if host := s.host; host != "" {
				muxRoute.Host(host)
			}
			if err := muxRoute.GetError(); err != nil {
				return nil, err
			}
			r.muxes = append(r.muxes, muxRoute)
			r.routes = append(r.routes, &routers.Route{
				Swagger:   doc,
				Server:    s.server,
				Path:      path,
				PathItem:  pathItem,
				Method:    "",
				Operation: nil,
			})
		}
	}
	return r, nil
}

// FindRoute extracts the route and parameters of an http.Request
func (r *Router) FindRoute(req *http.Request) (*routers.Route, map[string]string, error) {
	for i, muxRoute := range r.muxes {
		var match mux.RouteMatch
		if muxRoute.Match(req, &match) {
			if err := match.MatchErr; err != nil {
				// What then?
			}
			route := r.routes[i]
			route.Method = req.Method
			route.Operation = route.Swagger.Paths[route.Path].GetOperation(route.Method)
			return route, match.Vars, nil
		}
		switch match.MatchErr {
		case nil:
		case mux.ErrMethodMismatch:
			return nil, nil, routers.ErrMethodNotAllowed
		default: // What then?
		}
	}
	return nil, nil, routers.ErrPathNotFound
}

func orderedPaths(paths map[string]*openapi3.PathItem) []string {
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#pathsObject
	// When matching URLs, concrete (non-templated) paths would be matched
	// before their templated counterparts.
	// NOTE: sorting by number of variables ASC then by lexicographical
	// order seems to be a good heuristic.
	vars := make(map[int][]string)
	max := 0
	for path := range paths {
		count := strings.Count(path, "}")
		vars[count] = append(vars[count], path)
		if count > max {
			max = count
		}
	}
	ordered := make([]string, 0, len(paths))
	for c := 0; c <= max; c++ {
		if ps, ok := vars[c]; ok {
			sort.Strings(ps)
			for _, p := range ps {
				ordered = append(ordered, p)
			}
		}
	}
	return ordered
}

// Magic strings that temporarily replace "{}" so net/url.Parse() works
var blURL, brURL = strings.Repeat("-", 50), strings.Repeat("_", 50)

func bEncode(s string) string {
	s = strings.Replace(s, "{", blURL, -1)
	s = strings.Replace(s, "}", brURL, -1)
	return s
}
func bDecode(s string) string {
	s = strings.Replace(s, blURL, "{", -1)
	s = strings.Replace(s, brURL, "}", -1)
	return s
}
