package controller

type Config struct {
	ManifestsDir string     `yaml:"manifests_dir"`
	Manifests    []Manifest `yaml:"manifests"`
}

type Manifest struct {
	Name         string `yaml:"name"`
	TemplatePath string `yaml:"template_path"`
	// Metadata        Metadata `yaml:"metadata"`
	Args []string `yaml:"args"`
	// Image           string   `yaml:"image"`
	// Resources       Resource `yaml:"resources"`
	// StartupProbe    Probe    `yaml:"startupProbe"`
	// LivenessProbe   Probe    `yaml:"livenessProbe"`
	// SecurityContext Context  `yaml:"securityContext"`
}
