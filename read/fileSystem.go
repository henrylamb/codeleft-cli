package read

import (
	"os"
	"path/filepath"
)

// FileSystem abstracts file system operations.
type IFileSystem interface {
	Getwd() (string, error)
	Open(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
	Join(elem ...string) string
}

// OSFileSystem implements the FileSystem interface using the os package.
type OSFileSystem struct{}

func NewOSFileSystem() IFileSystem {
	return &OSFileSystem{}
}

func (o *OSFileSystem) Getwd() (string, error) {
	return os.Getwd()
}

func (o *OSFileSystem) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (o *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (o *OSFileSystem) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// JSONDecoder abstracts JSON decoding.
type JSONDecoder interface {
	Decode(v interface{}) error
}