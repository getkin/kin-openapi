package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/invopop/yaml"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	defaultDefaults = true
	defaults        = flag.Bool("defaults", defaultDefaults, "when false, disables schemas' default field validation")
)

var (
	defaultExamples = true
	examples        = flag.Bool("examples", defaultExamples, "when false, disables all example schema validation")
)

var (
	defaultExt = false
	ext        = flag.Bool("ext", defaultExt, "enables visiting other files")
)

var (
	defaultPatterns = true
	patterns        = flag.Bool("patterns", defaultPatterns, "when false, allows schema patterns unsupported by the Go regexp engine")
)

func main() {
	flag.Parse()
	filename := flag.Arg(0)
	if len(flag.Args()) != 1 || filename == "" {
		log.Fatalf("Usage: go run github.com/getkin/kin-openapi/cmd/validate@latest [--defaults] [--examples] [--ext] [--patterns] -- <local YAML or JSON file>\nGot: %+v\n", os.Args)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	var vd struct {
		OpenAPI string `json:"openapi" yaml:"openapi"`
		Swagger string `json:"swagger" yaml:"swagger"`
	}
	if err := yaml.Unmarshal(data, &vd); err != nil {
		log.Fatal(err)
	}

	switch {
	case vd.OpenAPI == "3" || strings.HasPrefix(vd.OpenAPI, "3."):
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = *ext

		doc, err := loader.LoadFromFile(filename)
		if err != nil {
			log.Fatalln("Loading error:", err)
		}

		var opts []openapi3.ValidationOption
		if !*defaults {
			opts = append(opts, openapi3.DisableSchemaDefaultsValidation())
		}
		if !*examples {
			opts = append(opts, openapi3.DisableExamplesValidation())
		}
		if !*patterns {
			opts = append(opts, openapi3.DisableSchemaPatternValidation())
		}

		if err = doc.Validate(loader.Context, opts...); err != nil {
			log.Fatalln("Validation error:", err)
		}

	case vd.Swagger == "2" || strings.HasPrefix(vd.Swagger, "2."):
		if *defaults != defaultDefaults {
			log.Fatal("Flag --defaults is only for OpenAPIv3")
		}
		if *examples != defaultExamples {
			log.Fatal("Flag --examples is only for OpenAPIv3")
		}
		if *ext != defaultExt {
			log.Fatal("Flag --ext is only for OpenAPIv3")
		}
		if *patterns != defaultPatterns {
			log.Fatal("Flag --patterns is only for OpenAPIv3")
		}

		var doc openapi2.T
		if err := yaml.Unmarshal(data, &doc); err != nil {
			log.Fatalln("Loading error:", err)
		}

	default:
		log.Fatal("Missing or incorrect 'openapi' or 'swagger' field")
	}
}
