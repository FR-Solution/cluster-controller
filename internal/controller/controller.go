package controller

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/fraima/cluster-controller/internal/utils"
	"gopkg.in/yaml.v3"
)

type controller struct {
	manifestDir string
	Values      map[string]interface{}
}

func New(cfg Config) (*controller, error) {
	s := &controller{
		manifestDir: cfg.ManifestsDir,
	}

	data, err := ioutil.ReadFile(cfg.BaseValuesFile)
	if err != nil {
		return nil, fmt.Errorf("read base values path: %s : %w", cfg.BaseValuesFile, err)
	}

	if err := yaml.Unmarshal(data, &s.Values); err != nil {
		return nil, fmt.Errorf("unmarshal base values: %w", err)
	}

	s.MergeValues(cfg.ExtraValues)

	for _, m := range cfg.Manifests {
		if s.manifestIsNotExist(m.Name) {
			if err := s.createManifest(m); err != nil {
				return nil, fmt.Errorf("create manifest %s : %w", m.Name, err)
			}
		}
	}

	return s, nil
}

func (s *controller) MergeValues(extraValues map[string]interface{}) {
	utils.MergeValues(s.Values, extraValues)
}

func (s *controller) createManifest(m Manifest) error {
	templateData, err := os.ReadFile(m.TemplatePath)
	if err != nil {
		return fmt.Errorf("open template file %s : %w", m.TemplatePath, err)
	}

	manifestTemplate, err := template.New(m.Name).Funcs(funcMap()).Parse(string(templateData))
	if err != nil {
		return fmt.Errorf("parse template %s : %w", m.TemplatePath, err)
	}

	var manifestBuffer bytes.Buffer
	if err = manifestTemplate.Execute(&manifestBuffer, s); err != nil {
		return fmt.Errorf("fill in template %s with data %+v : %w", m.TemplatePath, m, err)
	}

	if err = s.saveManifest(m.Name, manifestBuffer.Bytes()); err != nil {
		return fmt.Errorf("save manifest %s: %w", m.Name, err)
	}
	return nil
}
