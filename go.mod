module github.com/getkin/kin-openapi

go 1.14

replace github.com/ghodss/yaml/v2 => github.com/diamondburned/yaml/v2 v2.0.0-20240812065612-baf990d70122

require (
	github.com/ghodss/yaml v1.0.0
	github.com/ghodss/yaml/v2 v2.0.0-00010101000000-000000000000
	github.com/go-openapi/jsonpointer v0.19.5
	github.com/gorilla/mux v1.8.0
	github.com/stretchr/testify v1.5.1
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
