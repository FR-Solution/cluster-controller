package controller

type Config struct {
	ManifestsDir string                 `yaml:"manifestsDir"`
	Manifests    []Manifest             `yaml:"manifests"`
	Values       map[string]interface{} `yaml:"values"`
}

type Manifest struct {
	Name         string `yaml:"name"`
	TemplatePath string `yaml:"templatePath"`
}
