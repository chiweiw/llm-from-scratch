package db

type GlobalSetting struct {
	ID        string `db:"id" json:"id"`
	Key       string `db:"key" json:"key"`
	Value     string `db:"value" json:"value"`
	CreatedAt int64  `db:"created_at" json:"createdAt"`
	UpdatedAt int64  `db:"updated_at" json:"updatedAt"`
}

type SystemDefault struct {
	ID        string `db:"id" json:"id"`
	Key       string `db:"key" json:"key"`
	Value     string `db:"value" json:"value"`
	CreatedAt int64  `db:"created_at" json:"createdAt"`
	UpdatedAt int64  `db:"updated_at" json:"updatedAt"`
}

type Environment struct {
	ID            string `db:"id" json:"id"`
	Name          string `db:"name" json:"name"`
	Identifier    string `db:"identifier" json:"identifier"`
	Description   string `db:"description" json:"description"`
	ProjectRoot   string `db:"project_root" json:"projectRoot"`
	CloudDeploy   bool   `db:"cloud_deploy" json:"cloudDeploy"`
	Timeout       int    `db:"timeout" json:"timeout"`
	BackupCleanup bool   `db:"backup_cleanup" json:"backupCleanup"`
	CheckStatus   string `db:"check_status" json:"checkStatus"`
	CreatedAt     int64  `db:"created_at" json:"createdAt"`
	UpdatedAt     int64  `db:"updated_at" json:"updatedAt"`
}

type ServerConfig struct {
	ID            string `db:"id" json:"id"`
	EnvironmentID string `db:"environment_id" json:"environmentId"`
	Name          string `db:"name" json:"name"`
	Host          string `db:"host" json:"host"`
	Port          int    `db:"port" json:"port"`
	Username      string `db:"username" json:"username"`
	Password      string `db:"password" json:"password"`
	DeployDir     string `db:"deploy_dir" json:"deployDir"`
	RestartScript string `db:"restart_script" json:"restartScript"`
	EnableRestart bool   `db:"enable_restart" json:"enableRestart"`
	UseSudo       bool   `db:"use_sudo" json:"useSudo"`
	CreatedAt     int64  `db:"created_at" json:"createdAt"`
	UpdatedAt     int64  `db:"updated_at" json:"updatedAt"`
}

type TargetFile struct {
	ID            string `db:"id" json:"id"`
	EnvironmentID string `db:"environment_id" json:"environmentId"`
	LocalPath     string `db:"local_path" json:"localPath"`
	RemoteName    string `db:"remote_name" json:"remoteName"`
	DefaultCheck  bool   `db:"default_check" json:"defaultCheck"`
	CreatedAt     int64  `db:"created_at" json:"createdAt"`
	UpdatedAt     int64  `db:"updated_at" json:"updatedAt"`
}

type DeployHistory struct {
	ID              string `db:"id" json:"id"`
	EnvironmentID   string `db:"environment_id" json:"environmentId"`
	EnvironmentName string `db:"environment_name" json:"environmentName"`
	StartTime       int64  `db:"start_time" json:"startTime"`
	EndTime         int64  `db:"end_time" json:"endTime"`
	Status          string `db:"status" json:"status"`
	Files           string `db:"files" json:"files"`
	Duration        int64  `db:"duration" json:"duration"`
	ErrorMessage    string `db:"error_message" json:"errorMessage"`
	CreatedAt       int64  `db:"created_at" json:"createdAt"`
}

type FrontendDeployment struct {
	ID            string `db:"id" json:"id"`
	Name          string `db:"name" json:"name"`
	ProjectRoot   string `db:"project_root" json:"projectRoot"`
	BuildCommand  string `db:"build_command" json:"buildCommand"`
	BuildArgs     string `db:"build_args" json:"buildArgs"`
	OutputDir     string `db:"output_dir" json:"outputDir"`
	EnvironmentID string `db:"environment_id" json:"environmentId"`
	Enabled       bool   `db:"enabled" json:"enabled"`
	CreatedAt     int64  `db:"created_at" json:"createdAt"`
	UpdatedAt     int64  `db:"updated_at" json:"updatedAt"`
}

type DeployLog struct {
	ID        string `db:"id" json:"id"`
	DeployID  string `db:"deploy_id" json:"deployId"`
	Level     string `db:"level" json:"level"`
	Message   string `db:"message" json:"message"`
	Timestamp int64  `db:"timestamp" json:"timestamp"`
	CreatedAt int64  `db:"created_at" json:"createdAt"`
}

type Migration struct {
	ID        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	AppliedAt int64  `db:"applied_at" json:"appliedAt"`
}
