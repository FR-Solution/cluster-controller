package controller

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"text/template"
)

type controller struct {
	manifestDir string
}

func New(cfg Config) (*controller, error) {
	s := &controller{
		manifestDir: cfg.ManifestsDir,
	}

	for _, m := range cfg.Manifests {
		if s.manifestIsNotExist(m.Name) {
			if err := s.createManifest(m); err != nil {
				return nil, fmt.Errorf("create manifest %s : %w", m.Name, err)
			}
		}
	}

	return s, nil
}

func (s *controller) createManifest(i Manifest) error {
	templateData, err := os.ReadFile(i.TemplatePath)
	if err != nil {
		return fmt.Errorf("open template file %s : %w", i.TemplatePath, err)
	}

	manifestTemplate, err := template.New(i.Name).Parse(string(templateData))
	if err != nil {
		return fmt.Errorf("parse template %s : %w", i.TemplatePath, err)
	}

	var manifestBuffer bytes.Buffer
	if err = manifestTemplate.Execute(&manifestBuffer, i); err != nil {
		return fmt.Errorf("fill in template %s with data %+v : %w", i.TemplatePath, i, err)
	}

	if err = s.saveManifest(i.Name, manifestBuffer.Bytes()); err != nil {
		return fmt.Errorf("save manifest %s: %w", i.Name, err)
	}
	return nil
}

func (s *controller) saveManifest(name string, data []byte) error {
	path := path.Join(s.manifestDir, name+".yaml")
	return os.WriteFile(path, data, 0666)
}

func (s *controller) manifestIsNotExist(name string) bool {
	path := path.Join(s.manifestDir, name+".yaml")
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
