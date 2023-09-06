module github.com/getkin/kin-openapi

go 1.20

exclude (
	// these versions contain a nil pointer CVE
	gopkg.in/yaml.v3 v3.0.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

require (
	github.com/go-openapi/jsonpointer v0.19.6
	github.com/gorilla/mux v1.8.0
	github.com/invopop/yaml v0.2.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/perimeterx/marshmallow v1.1.5
	github.com/stretchr/testify v1.8.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)
