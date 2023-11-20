package openapi3

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	versionRegex = regexp.MustCompile(`^3\.\d+\.\d+$`)

	// ErrInvalidVersion is used when version is invalid (not 3.x.y)
	ErrInvalidVersion = errors.New("must be 3.x.y")
)

// Version is specified by Version/Swagger standard version 3.
// must be a sring
type Version string

// Validate returns an error if Schema does not comply with the Version spec.
func (oai Version) Validate(ctx context.Context) error {
	if vo := getValidationOptions(ctx); !vo.versionValidationDisabled {
		if !versionRegex.MatchString(string(oai)) {
			return ErrInvalidVersion
		}
	}
	return nil
}

// MinorVersion returns minor version from string assuming 0 is the default
// It is meaningful if and only if version vas validated
func (oai Version) Minor() uint64 {
	versionNumStrs := strings.Split(string(oai), ".")
	if len(versionNumStrs) > 1 {
		versionNum, err := strconv.ParseUint(versionNumStrs[1], 10, 64)
		if err == nil {
			return versionNum
		}
	}
	return 0
}
