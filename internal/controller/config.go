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

type Values struct {
	Global map[string]interface{} `yaml:"global"`
	// Container struct {
	// ETCD container `yaml:"etcd"`
	// } `yaml:"container"`
}

type global struct {
	Hostname    string `yaml:"hostname"`
	ClusterName string `yaml:"clusterName"`
	BaseDomain  string `yaml:"baseDomain"`
}

// type container struct {
// 	Args  []string `yaml:"args"`
// 	Image string   `yaml:"image"`
// }
