package dao

import (
	"deploy-tool/internal/config"
	"deploy-tool/internal/logger"
	"deploy-tool/internal/model/entity"
	"encoding/json"
	"os"
	"path/filepath"
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
	paths := []string{d.cfg.ConfigFilePath}
	if !filepath.IsAbs(d.cfg.ConfigFilePath) {
		if exe, err := os.Executable(); err == nil {
			paths = append(paths, filepath.Join(filepath.Dir(exe), d.cfg.ConfigFilePath))
		}
		paths = append(paths, filepath.Join(".", d.cfg.ConfigFilePath))
	}

	var data []byte
	var err error
	var usedPath string
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			usedPath = p
			break
		}
	}
	if err != nil {
		logger.Warn("未找到外置配置文件，回退到内置默认配置 (paths=%v)", paths)
		data = config.EmbeddedDefaultConfig()
		usedPath = "embedded"
	}
	logger.Info("读取配置文件成功: %s", usedPath)

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
			logger.Error("解析环境配置失败: %v", err)
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
