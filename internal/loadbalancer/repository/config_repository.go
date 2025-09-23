package repository

import (
	"GoRelay/pkg/logger"
	"GoRelay/pkg/utils"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigRepository struct {
	filePath string
	logger   *logger.Logger
}

func NewConfigRepository(filePath string, logger *logger.Logger) *ConfigRepository {
	return &ConfigRepository{
		filePath: filePath,
		logger:   logger,
	}
}

func (r *ConfigRepository) Load() (*utils.Config, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		logger.NewLogger().Error("error while reading config file", "error", err)
		return nil, err
	}
	var cfg utils.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		logger.NewLogger().Error("error while unmarshalling config-yaml.yaml", "error", err)
		return nil, err
	}

	err = utils.ValidateConfig(&cfg)
	if err != nil {
		logger.NewLogger().Error("error while valdiating config-yaml.yaml", "error", err)
	}
	return &cfg, nil
}
