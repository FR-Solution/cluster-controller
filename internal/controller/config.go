package controller

type Config struct {
	ManifestsPath string     `json:"manifests_path"`
	Manifests     []manifest `json:"manifests"`
}

type manifest struct {
	Name         string   `json:"name"`
	TemplatePath string   `json:"template_path"`
	Args         []string `json:"args"`
}
