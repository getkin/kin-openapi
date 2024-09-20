module github.com/getkin/kin-openapi

go 1.22.5

replace gopkg.in/yaml.v3 => github.com/oasdiff/yaml3 v0.0.0-20240920135353-c185dc6ea7c6

replace github.com/invopop/yaml => github.com/oasdiff/yaml v0.0.0-20240920191703-3e5a9fb5bdf3

require (
	github.com/go-openapi/jsonpointer v0.21.0
	github.com/gorilla/mux v1.8.0
	github.com/invopop/yaml v0.3.1
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/perimeterx/marshmallow v1.1.5
	github.com/stretchr/testify v1.9.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
)
