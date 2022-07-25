package controller

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"text/template"

	"go.uber.org/zap"
)

func (s *controller) prepareManifest(m Manifest) ([]byte, error) {
	if s.manifestIsExist(m.Name) {
		data, err := s.readManifest(m.Name)
		if err != nil {
			return nil, fmt.Errorf("read existing manifest: %w", err)
		}
		return data, nil
	}
	return s.RenderManifest(m)
}

func (s *controller) manifestIsExist(name string) bool {
	path := path.Join(s.manifestDir, name+".yaml")
	_, err := os.Stat(path)
	return os.IsExist(err)
}

func (s *controller) readManifest(name string) ([]byte, error) {
	path := path.Join(s.manifestDir, name+".yaml")
	return os.ReadFile(path)
}

func (s *controller) RenderManifest(m Manifest) ([]byte, error) {
	templateData, err := os.ReadFile(m.TemplatePath)
	if err != nil {
		return nil, fmt.Errorf("open template file %s : %w", m.TemplatePath, err)
	}

	manifestTemplate, err := template.New(m.Name).Funcs(funcMap()).Parse(string(templateData))
	if err != nil {
		return nil, fmt.Errorf("parse template %s : %w", m.TemplatePath, err)
	}

	var manifestBuffer bytes.Buffer
	if err = manifestTemplate.Execute(&manifestBuffer, s); err != nil {
		return nil, fmt.Errorf("fill in template %s with data %+v : %w", m.TemplatePath, m, err)
	}

	if err = s.saveManifest(m.Name, manifestBuffer.Bytes()); err != nil {
		return nil, fmt.Errorf("save manifest %s: %w", m.Name, err)
	}
	zap.L().Debug("manifest local save", zap.String("name", m.Name))

	return manifestBuffer.Bytes(), nil
}

func (s *controller) saveManifest(name string, data []byte) error {
	path := path.Join(s.manifestDir, name+".yaml")
	return os.WriteFile(path, data, 0666)
}
