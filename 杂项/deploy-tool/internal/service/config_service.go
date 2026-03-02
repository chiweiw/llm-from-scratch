package service

import (
	"deploy-tool/internal/dao"
	"deploy-tool/internal/model/entity"
	"deploy-tool/pkg/utils"
	"sync"
	"time"
)

type ConfigService struct {
	dao dao.ConfigDAO
	mu  sync.RWMutex
	cfg *entity.AppConfig
}

func NewConfigService(dao dao.ConfigDAO) *ConfigService {
	return &ConfigService{dao: dao}
}

func (s *ConfigService) Load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.dao.Load()
	if err != nil || cfg == nil {
		s.cfg = defaultConfig()
		_ = s.dao.Save(s.cfg)
		return
	}
	s.cfg = cfg
	_ = s.dao.Save(s.cfg)
}

func (s *ConfigService) Save() error {
	s.mu.RLock()
	cfg := s.cfg
	s.mu.RUnlock()
	if cfg == nil {
		return nil
	}
	return s.dao.Save(cfg)
}

func (s *ConfigService) Get() *entity.AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *ConfigService) GetSettings() entity.GlobalSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cfg == nil {
		return entity.GlobalSettings{}
	}
	return s.cfg.Settings
}

func (s *ConfigService) SaveSettings(settings entity.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg == nil {
		s.cfg = defaultConfig()
	}
	s.cfg.Settings = settings
	return s.dao.Save(s.cfg)
}

func (s *ConfigService) GetSystemDefaults() entity.SystemDefaultConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cfg == nil {
		return entity.SystemDefaultConfig{}
	}
	return s.cfg.SystemDefaults
}

func (s *ConfigService) SaveSystemDefaults(defaults entity.SystemDefaultConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg == nil {
		s.cfg = defaultConfig()
	}
	s.cfg.SystemDefaults = defaults
	return s.dao.Save(s.cfg)
}

func (s *ConfigService) GetEnvironments() []entity.Environment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cfg == nil {
		return nil
	}
	return s.cfg.Environments
}

func (s *ConfigService) GetEnvironment(id string) *entity.Environment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cfg == nil {
		return nil
	}
	for i := range s.cfg.Environments {
		if s.cfg.Environments[i].ID == id {
			env := s.cfg.Environments[i]
			return &env
		}
	}
	return nil
}

func (s *ConfigService) UpsertEnvironment(env entity.Environment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg == nil {
		s.cfg = defaultConfig()
	}

	now := time.Now().Unix()
	env.UpdatedAt = now

	found := false
	for i := range s.cfg.Environments {
		if s.cfg.Environments[i].ID == env.ID {
			s.cfg.Environments[i] = env
			found = true
			break
		}
	}
	if !found {
		if env.ID == "" {
			env.ID = "env_" + time.Now().Format("20060102150405")
		}
		env.CreatedAt = now
		s.cfg.Environments = append(s.cfg.Environments, env)
	}

	return s.dao.Save(s.cfg)
}

func (s *ConfigService) DeleteEnvironment(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg == nil {
		return nil
	}
	for i := range s.cfg.Environments {
		if s.cfg.Environments[i].ID == id {
			s.cfg.Environments = append(s.cfg.Environments[:i], s.cfg.Environments[i+1:]...)
			return s.dao.Save(s.cfg)
		}
	}
	return nil
}

func (s *ConfigService) ExportEnvironment(id string) (string, error) {
	env := s.GetEnvironment(id)
	if env == nil {
		return "", nil
	}
	data, err := utils.MarshalIndent(env)
	return string(data), err
}

func (s *ConfigService) ImportEnvironment(jsonData string) error {
	var env entity.Environment
	if err := utils.Unmarshal([]byte(jsonData), &env); err != nil {
		return err
	}
	env.ID = "env_" + time.Now().Format("20060102150405")
	return s.UpsertEnvironment(env)
}

func defaultConfig() *entity.AppConfig {
	now := time.Now().Unix()
	return &entity.AppConfig{
		Settings: entity.GlobalSettings{
			DefaultTimeout:   600,
			LogRetentionDays: 30,
			BackupEnabled:    true,
			NotifyOnComplete: true,
			CloudDeploy:      true,
			Theme:            "system",
			Language:         "zh-Hans",
		},
		SystemDefaults: entity.SystemDefaultConfig{
			JdkPath:           "",
			MavenPath:         "",
			MavenSettingsPath: "",
			MavenRepoPath:     "",
			MavenArgs:         []string{},
		},
		Environments: []entity.Environment{
			{
				ID:            "env_dev",
				Name:          "开发环境",
				Identifier:    "dev",
				Description:   "本地开发环境 - deploy_tool.py 配置",
				ProjectRoot:   `D:\javaproject\backcode`,
				CloudDeploy:   true,
				Timeout:       600,
				BackupCleanup: true,
				Servers:       []entity.ServerConfig{},
				TargetFiles:   []entity.TargetFile{},
				CheckStatus:   "unchecked",
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		},
		History: []entity.DeployHistory{},
	}
}
