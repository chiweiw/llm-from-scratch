package service

import (
	"deploy-tool/internal/db"
	"deploy-tool/internal/model/entity"
	"deploy-tool/pkg/utils"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type ConfigService struct {
	envDAO          db.EnvironmentDAO
	globalSettingDAO db.GlobalSettingDAO
	systemDefaultDAO db.SystemDefaultDAO
	serverConfigDAO db.ServerConfigDAO
	targetFileDAO   db.TargetFileDAO
	mu              sync.RWMutex
	cfg             *entity.AppConfig
}

func NewConfigService(
	envDAO db.EnvironmentDAO,
	globalSettingDAO db.GlobalSettingDAO,
	systemDefaultDAO db.SystemDefaultDAO,
	serverConfigDAO db.ServerConfigDAO,
	targetFileDAO db.TargetFileDAO,
) *ConfigService {
	return &ConfigService{
		envDAO:          envDAO,
		globalSettingDAO: globalSettingDAO,
		systemDefaultDAO: systemDefaultDAO,
		serverConfigDAO: serverConfigDAO,
		targetFileDAO:   targetFileDAO,
	}
}

func (s *ConfigService) Load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cfg = &entity.AppConfig{
		Settings:       s.loadSettings(),
		SystemDefaults: s.loadSystemDefaults(),
		Environments:   s.loadEnvironments(),
		History:        []entity.DeployHistory{},
	}

	if len(s.cfg.Environments) == 0 {
		s.createDefaultEnvironment()
	}
}

func (s *ConfigService) Save() error {
	s.mu.RLock()
	cfg := s.cfg
	s.mu.RUnlock()
	if cfg == nil {
		return nil
	}

	if err := s.saveSettings(cfg.Settings); err != nil {
		return err
	}

	if err := s.saveSystemDefaults(cfg.SystemDefaults); err != nil {
		return err
	}

	return nil
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
		return defaultSettings()
	}
	return s.cfg.Settings
}

func (s *ConfigService) SaveSettings(settings entity.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg == nil {
		s.cfg = &entity.AppConfig{}
	}
	s.cfg.Settings = settings
	return s.saveSettings(settings)
}

func (s *ConfigService) GetSystemDefaults() entity.SystemDefaultConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cfg == nil {
		return defaultSystemDefaults()
	}
	return s.cfg.SystemDefaults
}

func (s *ConfigService) SaveSystemDefaults(defaults entity.SystemDefaultConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg == nil {
		s.cfg = &entity.AppConfig{}
	}
	s.cfg.SystemDefaults = defaults
	return s.saveSystemDefaults(defaults)
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
		s.cfg = &entity.AppConfig{}
	}

	now := time.Now().Unix()
	env.UpdatedAt = now

	dbEnv := &db.Environment{
		ID:            env.ID,
		Name:          env.Name,
		Identifier:    env.Identifier,
		Description:   env.Description,
		ProjectRoot:   env.ProjectRoot,
		CloudDeploy:   env.CloudDeploy,
		Timeout:       env.Timeout,
		BackupCleanup: env.BackupCleanup,
		CheckStatus:   env.CheckStatus,
		CreatedAt:     env.CreatedAt,
		UpdatedAt:     env.UpdatedAt,
	}

	if env.ID == "" {
		env.ID = "env_" + time.Now().Format("20060102150405")
		dbEnv.ID = env.ID
		dbEnv.CreatedAt = now
	}

	if err := s.envDAO.Create(dbEnv); err != nil {
		return fmt.Errorf("保存环境失败: %w", err)
	}

	for _, server := range env.Servers {
		dbServer := &db.ServerConfig{
			ID:            server.ID,
			EnvironmentID: env.ID,
			Name:          server.Name,
			Host:          server.Host,
			Port:          server.Port,
			Username:      server.Username,
			Password:      server.Password,
			DeployDir:     server.DeployDir,
			RestartScript: server.RestartScript,
			EnableRestart: server.EnableRestart,
			UseSudo:       server.UseSudo,
		}

		if server.ID == "" {
			dbServer.ID = "server_" + time.Now().Format("20060102150405")
		}

		if err := s.serverConfigDAO.Create(dbServer); err != nil {
			return fmt.Errorf("保存服务器配置失败: %w", err)
		}
	}

	for _, file := range env.TargetFiles {
		dbFile := &db.TargetFile{
			ID:            file.ID,
			EnvironmentID: env.ID,
			LocalPath:     file.LocalPath,
			RemoteName:    file.RemoteName,
			DefaultCheck:  file.DefaultCheck,
		}

		if file.ID == "" {
			dbFile.ID = "file_" + time.Now().Format("20060102150405")
		}

		if err := s.targetFileDAO.Create(dbFile); err != nil {
			return fmt.Errorf("保存目标文件失败: %w", err)
		}
	}

	s.cfg.Environments = s.loadEnvironments()
	return nil
}

func (s *ConfigService) DeleteEnvironment(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.envDAO.Delete(id); err != nil {
		return err
	}

	s.cfg.Environments = s.loadEnvironments()
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

func (s *ConfigService) loadSettings() entity.GlobalSettings {
	settingsMap, err := s.globalSettingDAO.GetMap()
	if err != nil {
		return defaultSettings()
	}

	settings := defaultSettings()
	if timeout, ok := settingsMap["default_timeout"]; ok {
		fmt.Sscanf(timeout, "%d", &settings.DefaultTimeout)
	}
	if retention, ok := settingsMap["log_retention_days"]; ok {
		fmt.Sscanf(retention, "%d", &settings.LogRetentionDays)
	}
	if backup, ok := settingsMap["backup_enabled"]; ok {
		settings.BackupEnabled = backup == "true"
	}
	if notify, ok := settingsMap["notify_on_complete"]; ok {
		settings.NotifyOnComplete = notify == "true"
	}
	if cloud, ok := settingsMap["cloud_deploy"]; ok {
		settings.CloudDeploy = cloud == "true"
	}
	if theme, ok := settingsMap["theme"]; ok {
		settings.Theme = theme
	}
	if lang, ok := settingsMap["language"]; ok {
		settings.Language = lang
	}

	return settings
}

func (s *ConfigService) saveSettings(settings entity.GlobalSettings) error {
	if err := s.globalSettingDAO.Set("default_timeout", fmt.Sprintf("%d", settings.DefaultTimeout)); err != nil {
		return err
	}
	if err := s.globalSettingDAO.Set("log_retention_days", fmt.Sprintf("%d", settings.LogRetentionDays)); err != nil {
		return err
	}
	if err := s.globalSettingDAO.Set("backup_enabled", fmt.Sprintf("%v", settings.BackupEnabled)); err != nil {
		return err
	}
	if err := s.globalSettingDAO.Set("notify_on_complete", fmt.Sprintf("%v", settings.NotifyOnComplete)); err != nil {
		return err
	}
	if err := s.globalSettingDAO.Set("cloud_deploy", fmt.Sprintf("%v", settings.CloudDeploy)); err != nil {
		return err
	}
	if err := s.globalSettingDAO.Set("theme", settings.Theme); err != nil {
		return err
	}
	if err := s.globalSettingDAO.Set("language", settings.Language); err != nil {
		return err
	}
	return nil
}

func (s *ConfigService) loadSystemDefaults() entity.SystemDefaultConfig {
	defaultsMap, err := s.systemDefaultDAO.GetMap()
	if err != nil {
		return defaultSystemDefaults()
	}

	defaults := defaultSystemDefaults()
	if jdkPath, ok := defaultsMap["jdk_path"]; ok {
		defaults.JdkPath = jdkPath
	}
	if mavenPath, ok := defaultsMap["maven_path"]; ok {
		defaults.MavenPath = mavenPath
	}
	if mavenSettings, ok := defaultsMap["maven_settings_path"]; ok {
		defaults.MavenSettingsPath = mavenSettings
	}
	if mavenRepo, ok := defaultsMap["maven_repo_path"]; ok {
		defaults.MavenRepoPath = mavenRepo
	}
	if mavenArgs, ok := defaultsMap["maven_args"]; ok {
		json.Unmarshal([]byte(mavenArgs), &defaults.MavenArgs)
	}

	return defaults
}

func (s *ConfigService) saveSystemDefaults(defaults entity.SystemDefaultConfig) error {
	if err := s.systemDefaultDAO.Set("jdk_path", defaults.JdkPath); err != nil {
		return err
	}
	if err := s.systemDefaultDAO.Set("maven_path", defaults.MavenPath); err != nil {
		return err
	}
	if err := s.systemDefaultDAO.Set("maven_settings_path", defaults.MavenSettingsPath); err != nil {
		return err
	}
	if err := s.systemDefaultDAO.Set("maven_repo_path", defaults.MavenRepoPath); err != nil {
		return err
	}
	argsJSON, _ := json.Marshal(defaults.MavenArgs)
	if err := s.systemDefaultDAO.Set("maven_args", string(argsJSON)); err != nil {
		return err
	}
	return nil
}

func (s *ConfigService) loadEnvironments() []entity.Environment {
	dbEnvs, err := s.envDAO.GetAll()
	if err != nil {
		return []entity.Environment{}
	}

	envs := make([]entity.Environment, 0, len(dbEnvs))
	for _, dbEnv := range dbEnvs {
		env := entity.Environment{
			ID:            dbEnv.ID,
			Name:          dbEnv.Name,
			Identifier:    dbEnv.Identifier,
			Description:   dbEnv.Description,
			ProjectRoot:   dbEnv.ProjectRoot,
			CloudDeploy:   dbEnv.CloudDeploy,
			Timeout:       dbEnv.Timeout,
			BackupCleanup: dbEnv.BackupCleanup,
			CheckStatus:   dbEnv.CheckStatus,
			CreatedAt:     dbEnv.CreatedAt,
			UpdatedAt:     dbEnv.UpdatedAt,
		}

		servers, _ := s.serverConfigDAO.GetByEnvironmentID(dbEnv.ID)
		for _, dbServer := range servers {
			env.Servers = append(env.Servers, entity.ServerConfig{
				ID:            dbServer.ID,
				EnvironmentID: dbServer.EnvironmentID,
				Name:          dbServer.Name,
				Host:          dbServer.Host,
				Port:          dbServer.Port,
				Username:      dbServer.Username,
				Password:      dbServer.Password,
				DeployDir:     dbServer.DeployDir,
				RestartScript: dbServer.RestartScript,
				EnableRestart: dbServer.EnableRestart,
				UseSudo:       dbServer.UseSudo,
			})
		}

		files, _ := s.targetFileDAO.GetByEnvironmentID(dbEnv.ID)
		for _, dbFile := range files {
			env.TargetFiles = append(env.TargetFiles, entity.TargetFile{
				ID:            dbFile.ID,
				EnvironmentID: dbFile.EnvironmentID,
				LocalPath:     dbFile.LocalPath,
				RemoteName:    dbFile.RemoteName,
				DefaultCheck:  dbFile.DefaultCheck,
			})
		}

		envs = append(envs, env)
	}

	return envs
}

func (s *ConfigService) createDefaultEnvironment() {
	now := time.Now().Unix()
	env := entity.Environment{
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
	}
	_ = s.UpsertEnvironment(env)
}

func defaultSettings() entity.GlobalSettings {
	return entity.GlobalSettings{
		DefaultTimeout:   600,
		LogRetentionDays: 30,
		BackupEnabled:    true,
		NotifyOnComplete: true,
		CloudDeploy:      true,
		Theme:            "system",
		Language:         "zh-Hans",
	}
}

func defaultSystemDefaults() entity.SystemDefaultConfig {
	return entity.SystemDefaultConfig{
		JdkPath:           "",
		MavenPath:         "",
		MavenSettingsPath: "",
		MavenRepoPath:     "",
		MavenArgs:         []string{},
	}
}
