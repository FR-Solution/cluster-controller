package controller

import (
	"fmt"
	"io/ioutil"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	k8sYaml "sigs.k8s.io/yaml"

	"github.com/fraima/cluster-controller/internal/utils"
)

type K8sClient interface {
	CreateCRD() error
	CreateStaticPod(data []byte) error
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
		if manifest[m.Name], err = s.prepareManifest(m); err != nil {
			return nil, fmt.Errorf("create manifest %s : %w", m.Name, err)
		}
		zap.L().Debug("manifest created", zap.String("name", m.Name))
	}

	for {
		// time.Sleep(10 * time.Second)
		if err := s.cli.CreateCRD(); err != nil {
			zap.L().Warn("create crd", zap.Error(err))
			continue
		}
		break
	}

	for manifestName, manifestYAMLData := range manifest {
		manifestJSONData, _ := k8sYaml.YAMLToJSON(manifestYAMLData)
		if err := s.cli.CreateStaticPod(manifestJSONData); err != nil {
			return nil, fmt.Errorf("create saticpod %s : %w", manifestName, err)
		}
		zap.L().Debug("staticpod creatd", zap.String("name", manifestName))
	}

	return s, nil
}

func (s *controller) MergeValues(extraValues map[string]interface{}) {
	utils.MergeValues(s.Values, extraValues)
}


