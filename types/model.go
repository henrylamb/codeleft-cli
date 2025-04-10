package types

// Config represents the structure of the config.json file.
type Config struct {
	Threshold string `json:"threshold"`
	Security  struct {
		Owasp bool `json:"owasp"`
		Cwe  bool `json:"cwe"`
	} `json:"security"`
	Quality struct {
		Solid     bool `json:"solid"`
		PrReady   bool `json:"prReady"`
		CleanCode bool `json:"cleanCode"`
		Complexity bool `json:"complexity"`
		ComplexityPro bool `json:"complexityPro"`
	} `json:"quality"`
	SafetyCritical struct {
		MisraCpp bool `json:"misraCpp"`
	} `json:"safetyCritical"`
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
