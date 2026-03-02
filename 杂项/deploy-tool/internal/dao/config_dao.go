package dao

import (
	"deploy-tool/internal/config"
	"deploy-tool/internal/model/entity"
	"encoding/json"
	"os"
)

type ConfigDAO interface {
	Load() (*entity.AppConfig, error)
	Save(cfg *entity.AppConfig) error
}

type FileConfigDAO struct {
	cfg config.Config
}

func NewFileConfigDAO(cfg config.Config) *FileConfigDAO {
	return &FileConfigDAO{cfg: cfg}
}

func (d *FileConfigDAO) Load() (*entity.AppConfig, error) {
	data, err := os.ReadFile(d.cfg.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	type rawConfig struct {
		Settings       entity.GlobalSettings      `json:"settings"`
		SystemDefaults entity.SystemDefaultConfig `json:"systemDefaults"`
		Environments   []json.RawMessage          `json:"environments"`
		History        []entity.DeployHistory     `json:"history"`
	}

	var raw rawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	cfg := &entity.AppConfig{
		Settings:       raw.Settings,
		SystemDefaults: raw.SystemDefaults,
		Environments:   make([]entity.Environment, 0),
		History:        raw.History,
	}

	type envCompat struct {
		ProjectRoot string `json:"projectRoot"`
		Local       *struct {
			ProjectRoot string `json:"projectRoot"`
		} `json:"local"`
	}

	for _, envRaw := range raw.Environments {
		var env entity.Environment
		if err := json.Unmarshal(envRaw, &env); err != nil {
			continue
		}
		var compat envCompat
		_ = json.Unmarshal(envRaw, &compat)

		if env.ProjectRoot == "" {
			if compat.ProjectRoot != "" {
				env.ProjectRoot = compat.ProjectRoot
			} else if compat.Local != nil {
				env.ProjectRoot = compat.Local.ProjectRoot
			}
		}

		cfg.Environments = append(cfg.Environments, env)
	}

	return cfg, nil
}

func (d *FileConfigDAO) Save(cfg *entity.AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(d.cfg.ConfigFilePath, data, 0644)
}

