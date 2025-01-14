package types

// Config represents the structure of the config.json file.
type Config struct {
	Threshold string `json:"threshold"`
	Security  struct {
		Owasp bool `json:"owasp"`
	} `json:"security"`
	Quality struct {
		Solid     bool `json:"solid"`
		LiteTest  bool `json:"liteTest"`
		PrReady   bool `json:"prReady"`
		CleanCode bool `json:"cleanCode"`
	} `json:"quality"`
	Ignore struct {
		Files   []File   `json:"files"`
		Folders []string `json:"folders"`
	} `json:"ignore"`
}

// File represents a file to be ignored in the config.
type File struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
