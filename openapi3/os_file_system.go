package openapi3

import (
	"net/http"
	"os"
)

type osFileSystem struct{}

var _ http.FileSystem = (*osFileSystem)(nil)

func (o osFileSystem) Open(name string) (http.File, error) {
	return os.Open(name)
}
