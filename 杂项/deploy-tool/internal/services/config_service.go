package services

import (
	"deploy-tool/internal/models"
	"encoding/json"
	"os"
	"sync"
	"time"
)

var (
	configFilePath = "deploy-tool-config.json"
)

type ConfigService struct {
	config models.AppConfig
	mu     sync.RWMutex
}

func NewConfigService() *ConfigService {
	return &ConfigService{
		config: models.AppConfig{
			Settings: models.GlobalSettings{
				DefaultTimeout:   600,
				LogRetentionDays: 30,
				BackupEnabled:    true,
				NotifyOnComplete: true,
				CloudDeploy:      true,
				Theme:            "system",
				Language:         "zh-Hans",
			},
			SystemDefaults: models.SystemDefaultConfig{
				JdkPath:           "",
				MavenSettingsPath: "",
				MavenRepoPath:     "",
				MavenArgs:         []string{},
			},
			Environments: make([]models.Environment, 0),
			History:      make([]models.DeployHistory, 0),
		},
	}
}

func (s *ConfigService) Load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		s.initDefaultEnvironments()
		return
	}

	type rawConfig struct {
		Settings       models.GlobalSettings      `json:"settings"`
		SystemDefaults models.SystemDefaultConfig `json:"systemDefaults"`
		Environments   []json.RawMessage          `json:"environments"`
		History        []models.DeployHistory     `json:"history"`
	}

	var raw rawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		s.initDefaultEnvironments()
		return
	}

	cfg := models.AppConfig{
		Settings:       raw.Settings,
		SystemDefaults: raw.SystemDefaults,
		Environments:   make([]models.Environment, 0),
		History:        raw.History,
	}

	type envCompat struct {
		ProjectRoot string `json:"projectRoot"`
		Local       *struct {
			ProjectRoot string `json:"projectRoot"`
		} `json:"local"`
	}

	for _, envRaw := range raw.Environments {
		var env models.Environment
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

	s.config = cfg
}

func (s *ConfigService) initDefaultEnvironments() {
	now := time.Now().Unix()
	s.config.Environments = []models.Environment{
		{
			ID:            "env_dev",
			Name:          "开发环境",
			Identifier:    "dev",
			Description:   "本地开发环境 - deploy_tool.py 配置",
			ProjectRoot:   `D:\javaproject\backcode`,
			CloudDeploy:   true,
			Timeout:       600,
			BackupCleanup: true,
			Servers: []models.ServerConfig{
				{
					ID:            "server_dev_1",
					Name:          "开发服务器 (192.168.8.26)",
					Host:          "192.168.8.26",
					Port:          22221,
					Username:      "omp",
					Password:      "cB7JzLsk",
					DeployDir:     "/home/omp/shanguotou/jar/",
					RestartScript: "/home/omp/shanguotou/jar/restart_jar_dev.sh",
					EnableRestart: true,
					UseSudo:       false,
				},
			},
			TargetFiles: []models.TargetFile{
				{
					ID:           "jar_1",
					LocalPath:    `startup\platform-startup-project\target\platform-startup-project.jar`,
					RemoteName:   "",
					DefaultCheck: true,
				},
				{
					ID:           "jar_2",
					LocalPath:    `startup\platform-startup-system\target\platform-startup-system.jar`,
					RemoteName:   "",
					DefaultCheck: true,
				},
				{
					ID:           "jar_3",
					LocalPath:    `startup\platform-startup-customer\target\platform-startup-customer.jar`,
					RemoteName:   "",
					DefaultCheck: false,
				},
			},
			CheckStatus: "unchecked",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

func (s *ConfigService) Save() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(configFilePath, data, 0644)
}

func (s *ConfigService) GetConfig() *models.AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config
}

func (s *ConfigService) GetEnvironments() []models.Environment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Environments
}

func (s *ConfigService) GetEnvironment(id string) *models.Environment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.config.Environments {
		if s.config.Environments[i].ID == id {
			return &s.config.Environments[i]
		}
	}
	return nil
}

func (s *ConfigService) SaveEnvironment(env models.Environment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	env.UpdatedAt = time.Now().Unix()
	found := false
	for i := range s.config.Environments {
		if s.config.Environments[i].ID == env.ID {
			s.config.Environments[i] = env
			found = true
			break
		}
	}
	if !found {
		env.CreatedAt = time.Now().Unix()
		s.config.Environments = append(s.config.Environments, env)
	}
	return nil
}

func (s *ConfigService) DeleteEnvironment(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.config.Environments {
		if s.config.Environments[i].ID == id {
			s.config.Environments = append(s.config.Environments[:i], s.config.Environments[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *ConfigService) Export(envID string) (string, error) {
	env := s.GetEnvironment(envID)
	if env == nil {
		return "", nil
	}
	data, err := json.MarshalIndent(env, "", "  ")
	return string(data), err
}

func (s *ConfigService) Import(jsonData string) error {
	var env models.Environment
	if err := json.Unmarshal([]byte(jsonData), &env); err != nil {
		return err
	}
	env.ID = "env_" + time.Now().Format("20060102150405")
	return s.SaveEnvironment(env)
}

func (s *ConfigService) GetSettings() *models.GlobalSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.Settings
}

func (s *ConfigService) SaveSettings(settings models.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.Settings = settings
	return nil
}

func (s *ConfigService) GetSystemDefaults() *models.SystemDefaultConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.config.SystemDefaults
}

func (s *ConfigService) SaveSystemDefaults(defaults models.SystemDefaultConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.SystemDefaults = defaults
	return nil
}
