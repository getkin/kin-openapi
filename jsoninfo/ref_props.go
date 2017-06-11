package jsoninfo

// RefProps provides support for JSON references
type RefProps struct {
	Ref            string `json:"$ref,omitempty"`
	IsUnmarshalled bool   `json:"-"` // If true, the object was unmarshalled from JSON
}

func (refProps *RefProps) GetRef() string {
	return refProps.Ref
}

func (refProps *RefProps) SetRef(value string) {
	refProps.Ref = value
}

type RefHolder interface {
	GetRef() string
	SetRef(value string)
}
