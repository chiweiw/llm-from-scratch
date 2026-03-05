package entity

type Environment struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Identifier  string         `json:"identifier"`
	Description string         `json:"description"`
	ProjectRoot string         `json:"projectRoot"`
	BuildType   string         `json:"buildType"`
	CloudDeploy bool           `json:"cloudDeploy"`
	Timeout     int            `json:"timeout"`
	Servers     []ServerConfig `json:"servers"`
	TargetFiles []TargetFile   `json:"targetFiles"`
	CheckStatus string         `json:"checkStatus"`
	CreatedAt   int64          `json:"createdAt"`
	UpdatedAt   int64          `json:"updatedAt"`
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
	UrlPath      string `json:"urlPath"`
	DefaultCheck bool   `json:"defaultCheck"`
}
