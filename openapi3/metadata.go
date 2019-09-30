package openapi3

import "net/url"

type Metadata struct {
	ID   string  `json:"-"`
	Path url.URL `json:"-"`
}
