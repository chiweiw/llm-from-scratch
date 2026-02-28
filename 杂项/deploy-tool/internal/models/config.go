package models

type GlobalSettings struct {
	DefaultTimeout   int    `json:"defaultTimeout"`
	LogRetentionDays int    `json:"logRetentionDays"`
	BackupEnabled    bool   `json:"backupEnabled"`
	NotifyOnComplete bool   `json:"notifyOnComplete"`
	CloudDeploy      bool   `json:"cloudDeploy"`
	Theme            string `json:"theme"`
	Language         string `json:"language"`
}

type SystemDefaultConfig struct {
	JdkPath           string `json:"jdkPath"`
	MavenPath         string `json:"mavenPath"`
	MavenSettingsPath string `json:"mavenSettingsPath"`
	MavenRepoPath     string `json:"mavenRepoPath"`
	MavenArgs         string `json:"mavenArgs"`
}

type AppConfig struct {
	Settings       GlobalSettings      `json:"settings"`
	SystemDefaults SystemDefaultConfig `json:"systemDefaults"`
	Environments   []Environment       `json:"environments"`
	History        []DeployHistory     `json:"history"`
}
