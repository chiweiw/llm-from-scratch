package models

type Environment struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Identifier    string         `json:"identifier"`
	Description   string         `json:"description"`
	CloudDeploy   bool           `json:"cloudDeploy"`
	Timeout       int            `json:"timeout"`
	DryRun        bool           `json:"dryRun"`
	BackupCleanup bool           `json:"backupCleanup"`
	Local         LocalConfig    `json:"local"`
	Servers       []ServerConfig `json:"servers"`
	TargetFiles   []TargetFile   `json:"targetFiles"`
	CheckStatus   string         `json:"checkStatus"`
	CreatedAt     int64          `json:"createdAt"`
	UpdatedAt     int64          `json:"updatedAt"`
}

type LocalConfig struct {
	ProjectRoot       string `json:"projectRoot"`
	JdkPath           string `json:"jdkPath"`
	MavenPath         string `json:"mavenPath"`
	MavenSettingsPath string `json:"mavenSettingsPath"`
	MavenRepoPath     string `json:"mavenRepoPath"`
	MavenArgs         string `json:"mavenArgs"`
	MavenQuiet        bool   `json:"mavenQuiet"`
	CompactMvnLog     bool   `json:"compactMvnLog"`
	SpecifyPom        bool   `json:"specifyPom"`
	OfflineBuild      bool   `json:"offlineBuild"`
}

type ServerConfig struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	DeployDir     string `json:"deployDir"`
	RestartScript string `json:"restartScript"`
	EnableRestart bool   `json:"enableRestart"`
	UseSudo       bool   `json:"useSudo"`
}

type TargetFile struct {
	ID           string `json:"id"`
	LocalPath    string `json:"localPath"`
	RemoteName   string `json:"remoteName"`
	DefaultCheck bool   `json:"defaultCheck"`
}
