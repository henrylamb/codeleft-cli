package read

import (
	"codeleft-cli/types"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileSystem abstracts file system operations.
type FileSystem interface {
	Getwd() (string, error)
	Open(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
	Join(elem ...string) string
}

// OSFileSystem implements the FileSystem interface using the os package.
type OSFileSystem struct{}

func NewOSFileSystem() FileSystem {
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

// ConfigSource interface for reading configuration.
type ConfigSource interface {
	ReadConfig() (*types.Config, error)
}

// ConfigPathResolver resolves the path to the config file.
type ConfigPathResolver interface {
	ResolveConfigPath() (string, error)
}

// ConfigJSONReader reads the config from a JSON file.
type ConfigJSONReader interface {
	ReadConfigFromJSON(path string) (*types.Config, error)
}

// ConfigReader is responsible for reading the config.json file.
type ConfigReader struct {
	RepoRoot     string
	CodeleftPath string
	FileSystem   FileSystem
	//	JSONDecoder  JSONDecoder // Removed direct dependency on JSONDecoder
}

// NewConfigReader creates a new ConfigReader.
func NewConfigReader(fs FileSystem) (*ConfigReader, error) {
	if fs == nil {
		fs = &OSFileSystem{} // Use OSFileSystem as default
	}
	repoRoot, err := fs.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Recursively find .codeleft
	codeleftPath, err := findCodeleftRecursive(repoRoot)
	if err != nil {
		return nil, err
	}

	cr := &ConfigReader{
		RepoRoot:     repoRoot,
		CodeleftPath: codeleftPath,
		FileSystem:   fs,
	}
	return cr, nil
}

// ResolveConfigPath resolves the path to the config.json file.
func (cr *ConfigReader) ResolveConfigPath() (string, error) {
	if cr.CodeleftPath == "" {
		return "", fmt.Errorf(".codeLeft folder not found in the repository root: %s", cr.RepoRoot)
	}
	return filepath.Join(cr.CodeleftPath, "config.json"), nil
}

// ReadConfig reads the configuration from config.json.
func (cr *ConfigReader) ReadConfig() (*types.Config, error) {
	configPath, err := cr.ResolveConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config.json exists
	info, err := cr.FileSystem.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config.json does not exist at path: %s", configPath)
		}
		return nil, fmt.Errorf("error accessing config.json: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("config.json exists but is a directory: %s", configPath)
	}

	// Open the config.json file
	file, err := cr.FileSystem.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w", err)
	}
	defer file.Close()

	// Decode the JSON into a Config struct
	var config types.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config.json: %w", err)
	}

	return &config, nil
}