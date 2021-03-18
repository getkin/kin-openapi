package openapi3filter

import (
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

// Route describes the operation an http.Request can match
type Route struct {
	Swagger   *openapi3.Swagger
	Server    *openapi3.Server
	Path      string
	PathItem  *openapi3.PathItem
	Method    string
	Operation *openapi3.Operation
}

// Router helps link http.Request.s and an OpenAPIv3 spec
type Router struct {
	muxes  []*mux.Route
	routes []*Route
}

// NewRouter creates a gorilla/mux router.
// Assumes spec is .Validate()d
// TODO: Handle/HandlerFunc + ServeHTTP (When there is a match, the route variables can be retrieved calling mux.Vars(request))
func NewRouter(doc *openapi3.Swagger) (*Router, error) {
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
		for method, op := range pathItem.Operations() {
			for _, s := range servers {
				muxRoute := muxRouter.Methods(method).Path(s.base + path)
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
				r.routes = append(r.routes, &Route{
					Swagger:   doc,
					Server:    s.server,
					Path:      path,
					PathItem:  pathItem,
					Method:    method,
					Operation: op,
				})
			}
		}
	}
	return r, nil
}

// FindRoute extracts the route and parameters of an http.Request
func (r *Router) FindRoute(req *http.Request) (*Route, map[string]string, error) {
	for i, muxRoute := range r.muxes {
		var match mux.RouteMatch
		if muxRoute.Match(req, &match) {
			if err := match.MatchErr; err != nil {
				return nil, nil, err
			}
			return r.routes[i], match.Vars, nil
		}
	}
	return nil, nil, ErrPathNotFound
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
