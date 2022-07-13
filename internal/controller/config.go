package controller

type Config struct {
	ManifestsDir   string                 `yaml:"manifestsDir"`
	Manifests      []Manifest             `yaml:"manifests"`
	BaseValuesFile string                 `yaml:"baseValuesFile"`
	ExtraValues    map[string]interface{} `yaml:"extraValues"`
}

type Manifest struct {
	Name         string `yaml:"name"`
	TemplatePath string `yaml:"templatePath"`
}
