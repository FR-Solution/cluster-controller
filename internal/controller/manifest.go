package controller

import (
	"os"
	"path"
)

func (s *controller) saveManifest(name string, data []byte) error {
	path := path.Join(s.manifestDir, name+".yaml")
	return os.WriteFile(path, data, 0666)
}

func (s *controller) manifestIsNotExist(name string) bool {
	path := path.Join(s.manifestDir, name+".yaml")
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
