package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/fraima/cluster-controller/internal/controller"
)

type Config struct {
	Controller controller.Config `yaml:",inline"`
}

// Read config by path.
func Read(path string) (cfg Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("read config file %s: %w", path, err)
		return
	}
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		err = fmt.Errorf("unmarshal config %w", err)
		return
	}
	return
}
