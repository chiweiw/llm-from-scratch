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
				JdkPath:           `C:\Program Files\Java\jdk1.8.0_202\bin`,
				MavenSettingsPath: `D:\java_tools\apache-maven-3.9.12\conf\settings_sgt0903.xml`,
				MavenRepoPath:     `D:\m2\repository`,
				MavenArgs:         "clean package -DskipTests",
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

	if err := json.Unmarshal(data, &s.config); err != nil {
		s.initDefaultEnvironments()
		return
	}
}

func (s *ConfigService) initDefaultEnvironments() {
	now := time.Now().Unix()
	s.config.Environments = []models.Environment{
		{
			ID:            "env_dev",
			Name:          "开发环境",
			Identifier:    "dev",
			Description:   "本地开发环境 - deploy_tool.py 配置",
			CloudDeploy:   true,
			Timeout:       600,
			DryRun:        false,
			BackupCleanup: true,
			Local: models.LocalConfig{
				ProjectRoot:       `D:\javaproject\backcode`,
				JdkPath:           `C:\Program Files\Java\jdk1.8.0_202\bin`,
				MavenSettingsPath: `D:\java_tools\apache-maven-3.9.12\conf\settings_sgt0903.xml`,
				MavenRepoPath:     `D:\m2\repository`,
				MavenArgs:         "clean package -DskipTests -s D:\\java_tools\\apache-maven-3.9.12\\conf\\settings_sgt0903.xml -Dmaven.repo.local=D:\\m2\\repository",
				MavenQuiet:        true,
				CompactMvnLog:     true,
				SpecifyPom:        true,
				OfflineBuild:      true,
			},
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
