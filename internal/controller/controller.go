package controller

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/fraima/cluster-controller/internal/utils"
)

type K8sClient interface {
	CreateCRD() error
}

type controller struct {
	cli K8sClient

	manifestDir string
	Values      map[string]interface{}
}

func New(cli K8sClient, cfg Config) (*controller, error) {
	s := &controller{
		cli:         cli,
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
	manifest := make(map[string][]byte, 0)

	for _, m := range cfg.Manifests {
		if s.manifestIsNotExist(m.Name) {
			if manifest[m.Name], err = s.createManifest(m); err != nil {
				return nil, fmt.Errorf("create manifest %s : %w", m.Name, err)
			}
		}
	}

	for {
		// time.Sleep(10 * time.Second)
		if err := s.cli.CreateCRD(); err != nil {
			zap.L().Warn("create crd", zap.Error(err))
			continue
		}
		break
	}

	return s, nil
}

func (s *controller) MergeValues(extraValues map[string]interface{}) {
	utils.MergeValues(s.Values, extraValues)
}

func (s *controller) createManifest(m Manifest) ([]byte, error) {
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
	zap.L().Debug("manifest create", zap.String("name", m.Name))
	return manifestBuffer.Bytes(), nil
}
