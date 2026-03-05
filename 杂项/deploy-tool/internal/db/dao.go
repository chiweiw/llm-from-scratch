package db

import (
	"database/sql"
	"fmt"
	"time"
)

type EnvironmentDAO interface {
	GetAll() ([]Environment, error)
	GetByID(id string) (*Environment, error)
	GetByIdentifier(identifier string) (*Environment, error)
	Create(env *Environment) error
	Update(env *Environment) error
	Delete(id string) error
}

type GlobalSettingDAO interface {
	GetAll() ([]GlobalSetting, error)
	GetByKey(key string) (*GlobalSetting, error)
	GetMap() (map[string]string, error)
	Set(key, value string) error
	Delete(key string) error
}

type SystemDefaultDAO interface {
	GetAll() ([]SystemDefault, error)
	GetByKey(key string) (*SystemDefault, error)
	GetMap() (map[string]string, error)
	Set(key, value string) error
	Delete(key string) error
}

type ServerConfigDAO interface {
	GetByEnvironmentID(envID string) ([]ServerConfig, error)
	GetByID(id string) (*ServerConfig, error)
	Create(config *ServerConfig) error
	Update(config *ServerConfig) error
	Delete(id string) error
}

type TargetFileDAO interface {
	GetByEnvironmentID(envID string) ([]TargetFile, error)
	GetByID(id string) (*TargetFile, error)
	Create(file *TargetFile) error
	Update(file *TargetFile) error
	Delete(id string) error
}

type DeployHistoryDAO interface {
	GetAll(limit int) ([]DeployHistory, error)
	GetByEnvironmentID(envID string, limit int) ([]DeployHistory, error)
	GetByID(id string) (*DeployHistory, error)
	Create(history *DeployHistory) error
	Update(history *DeployHistory) error
	Delete(id string) error
	DeleteOld(days int) error
}

type DeployLogDAO interface {
	GetByDeployID(deployID string) ([]DeployLog, error)
	Create(log *DeployLog) error
	BatchCreate(logs []DeployLog) error
	DeleteByDeployID(deployID string) error
}

type FrontendDeploymentDAO interface {
	GetAll() ([]FrontendDeployment, error)
	GetByID(id string) (*FrontendDeployment, error)
	GetByEnvironmentID(envID string) ([]FrontendDeployment, error)
	GetEnabled() ([]FrontendDeployment, error)
	Create(deployment *FrontendDeployment) error
	Update(deployment *FrontendDeployment) error
	Delete(id string) error
}

type environmentDAO struct {
	db *sql.DB
}

func NewEnvironmentDAO(database *Database) EnvironmentDAO {
	return &environmentDAO{db: database.db}
}

func (d *environmentDAO) GetAll() ([]Environment, error) {
	query := `SELECT id, name, identifier, description, project_root, cloud_deploy, 
			  timeout, check_status, created_at, updated_at 
			  FROM environments ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []Environment
	for rows.Next() {
		var env Environment
		err := rows.Scan(
			&env.ID, &env.Name, &env.Identifier, &env.Description, &env.ProjectRoot,
			&env.CloudDeploy, &env.Timeout, &env.CheckStatus,
			&env.CreatedAt, &env.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		envs = append(envs, env)
	}
	return envs, nil
}

func (d *environmentDAO) GetByID(id string) (*Environment, error) {
	query := `SELECT id, name, identifier, description, project_root, cloud_deploy, 
			  timeout, check_status, created_at, updated_at 
			  FROM environments WHERE id = ?`

	var env Environment
	err := d.db.QueryRow(query, id).Scan(
		&env.ID, &env.Name, &env.Identifier, &env.Description, &env.ProjectRoot,
		&env.CloudDeploy, &env.Timeout, &env.CheckStatus,
		&env.CreatedAt, &env.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (d *environmentDAO) GetByIdentifier(identifier string) (*Environment, error) {
	query := `SELECT id, name, identifier, description, project_root, cloud_deploy, 
			  timeout, check_status, created_at, updated_at 
			  FROM environments WHERE identifier = ?`

	var env Environment
	err := d.db.QueryRow(query, identifier).Scan(
		&env.ID, &env.Name, &env.Identifier, &env.Description, &env.ProjectRoot,
		&env.CloudDeploy, &env.Timeout, &env.CheckStatus,
		&env.CreatedAt, &env.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (d *environmentDAO) Create(env *Environment) error {
	query := `INSERT OR REPLACE INTO environments (id, name, identifier, description, project_root, 
			  cloud_deploy, timeout, check_status, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Unix()
	createdAt := env.CreatedAt
	if createdAt == 0 {
		createdAt = now
	}
	_, err := d.db.Exec(query, env.ID, env.Name, env.Identifier, env.Description,
		env.ProjectRoot, env.CloudDeploy, env.Timeout,
		env.CheckStatus, createdAt, now)
	return err
}

func (d *environmentDAO) Update(env *Environment) error {
	query := `UPDATE environments SET name = ?, identifier = ?, description = ?, 
			  project_root = ?, cloud_deploy = ?, timeout = ?, 
			  check_status = ?, updated_at = ? WHERE id = ?`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, env.Name, env.Identifier, env.Description,
		env.ProjectRoot, env.CloudDeploy, env.Timeout,
		env.CheckStatus, now, env.ID)
	return err
}

func (d *environmentDAO) Delete(id string) error {
	_, err := d.db.Exec("DELETE FROM environments WHERE id = ?", id)
	return err
}

type globalSettingDAO struct {
	db *sql.DB
}

func NewGlobalSettingDAO(database *Database) GlobalSettingDAO {
	return &globalSettingDAO{db: database.db}
}

func (d *globalSettingDAO) GetAll() ([]GlobalSetting, error) {
	rows, err := d.db.Query("SELECT id, key, value, created_at, updated_at FROM global_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []GlobalSetting
	for rows.Next() {
		var s GlobalSetting
		err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (d *globalSettingDAO) GetByKey(key string) (*GlobalSetting, error) {
	var s GlobalSetting
	err := d.db.QueryRow("SELECT id, key, value, created_at, updated_at FROM global_settings WHERE key = ?", key).
		Scan(&s.ID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *globalSettingDAO) GetMap() (map[string]string, error) {
	settings, err := d.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (d *globalSettingDAO) Set(key, value string) error {
	now := time.Now().Unix()

	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM global_settings WHERE key = ?)", key).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		_, err = d.db.Exec("UPDATE global_settings SET value = ?, updated_at = ? WHERE key = ?", value, now, key)
	} else {
		_, err = d.db.Exec("INSERT INTO global_settings (id, key, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			generateID(), key, value, now, now)
	}
	return err
}

func (d *globalSettingDAO) Delete(key string) error {
	_, err := d.db.Exec("DELETE FROM global_settings WHERE key = ?", key)
	return err
}

type systemDefaultDAO struct {
	db *sql.DB
}

func NewSystemDefaultDAO(database *Database) SystemDefaultDAO {
	return &systemDefaultDAO{db: database.db}
}

func (d *systemDefaultDAO) GetAll() ([]SystemDefault, error) {
	rows, err := d.db.Query("SELECT id, key, value, created_at, updated_at FROM system_defaults")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defaults []SystemDefault
	for rows.Next() {
		var s SystemDefault
		err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		defaults = append(defaults, s)
	}
	return defaults, nil
}

func (d *systemDefaultDAO) GetByKey(key string) (*SystemDefault, error) {
	var s SystemDefault
	err := d.db.QueryRow("SELECT id, key, value, created_at, updated_at FROM system_defaults WHERE key = ?", key).
		Scan(&s.ID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *systemDefaultDAO) GetMap() (map[string]string, error) {
	defaults, err := d.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range defaults {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (d *systemDefaultDAO) Set(key, value string) error {
	now := time.Now().Unix()

	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM system_defaults WHERE key = ?)", key).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		_, err = d.db.Exec("UPDATE system_defaults SET value = ?, updated_at = ? WHERE key = ?", value, now, key)
	} else {
		_, err = d.db.Exec("INSERT INTO system_defaults (id, key, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			generateID(), key, value, now, now)
	}
	return err
}

func (d *systemDefaultDAO) Delete(key string) error {
	_, err := d.db.Exec("DELETE FROM system_defaults WHERE key = ?", key)
	return err
}

type serverConfigDAO struct {
	db *sql.DB
}

func NewServerConfigDAO(database *Database) ServerConfigDAO {
	return &serverConfigDAO{db: database.db}
}

func (d *serverConfigDAO) GetByEnvironmentID(envID string) ([]ServerConfig, error) {
	query := `SELECT id, environment_id, name, host, port, username, password, 
			  deploy_dir, restart_script, enable_restart, use_sudo, created_at, updated_at 
			  FROM server_configs WHERE environment_id = ?`

	rows, err := d.db.Query(query, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []ServerConfig
	for rows.Next() {
		var cfg ServerConfig
		err := rows.Scan(
			&cfg.ID, &cfg.EnvironmentID, &cfg.Name, &cfg.Host, &cfg.Port,
			&cfg.Username, &cfg.Password, &cfg.DeployDir, &cfg.RestartScript,
			&cfg.EnableRestart, &cfg.UseSudo, &cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func (d *serverConfigDAO) GetByID(id string) (*ServerConfig, error) {
	query := `SELECT id, environment_id, name, host, port, username, password, 
			  deploy_dir, restart_script, enable_restart, use_sudo, created_at, updated_at 
			  FROM server_configs WHERE id = ?`

	var cfg ServerConfig
	err := d.db.QueryRow(query, id).Scan(
		&cfg.ID, &cfg.EnvironmentID, &cfg.Name, &cfg.Host, &cfg.Port,
		&cfg.Username, &cfg.Password, &cfg.DeployDir, &cfg.RestartScript,
		&cfg.EnableRestart, &cfg.UseSudo, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (d *serverConfigDAO) Create(config *ServerConfig) error {
	query := `INSERT OR REPLACE INTO server_configs (id, environment_id, name, host, port, username, 
			  password, deploy_dir, restart_script, enable_restart, use_sudo, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, config.ID, config.EnvironmentID, config.Name, config.Host,
		config.Port, config.Username, config.Password, config.DeployDir, config.RestartScript,
		config.EnableRestart, config.UseSudo, now, now)
	return err
}

func (d *serverConfigDAO) Update(config *ServerConfig) error {
	query := `UPDATE server_configs SET name = ?, host = ?, port = ?, username = ?, 
			  password = ?, deploy_dir = ?, restart_script = ?, enable_restart = ?, 
			  use_sudo = ?, updated_at = ? WHERE id = ?`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, config.Name, config.Host, config.Port, config.Username,
		config.Password, config.DeployDir, config.RestartScript, config.EnableRestart,
		config.UseSudo, now, config.ID)
	return err
}

func (d *serverConfigDAO) Delete(id string) error {
	_, err := d.db.Exec("DELETE FROM server_configs WHERE id = ?", id)
	return err
}

type targetFileDAO struct {
	db *sql.DB
}

func NewTargetFileDAO(database *Database) TargetFileDAO {
	return &targetFileDAO{db: database.db}
}

func (d *targetFileDAO) GetByEnvironmentID(envID string) ([]TargetFile, error) {
	query := `SELECT id, environment_id, local_path, remote_name, default_check, created_at, updated_at 
			  FROM target_files WHERE environment_id = ?`

	rows, err := d.db.Query(query, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []TargetFile
	for rows.Next() {
		var f TargetFile
		err := rows.Scan(&f.ID, &f.EnvironmentID, &f.LocalPath, &f.RemoteName, &f.DefaultCheck, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (d *targetFileDAO) GetByID(id string) (*TargetFile, error) {
	query := `SELECT id, environment_id, local_path, remote_name, default_check, created_at, updated_at 
			  FROM target_files WHERE id = ?`

	var f TargetFile
	err := d.db.QueryRow(query, id).Scan(&f.ID, &f.EnvironmentID, &f.LocalPath, &f.RemoteName, &f.DefaultCheck, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (d *targetFileDAO) Create(file *TargetFile) error {
	query := `INSERT OR REPLACE INTO target_files (id, environment_id, local_path, remote_name, default_check, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, file.ID, file.EnvironmentID, file.LocalPath, file.RemoteName, file.DefaultCheck, now, now)
	return err
}

func (d *targetFileDAO) Update(file *TargetFile) error {
	query := `UPDATE target_files SET local_path = ?, remote_name = ?, default_check = ?, updated_at = ? WHERE id = ?`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, file.LocalPath, file.RemoteName, file.DefaultCheck, now, file.ID)
	return err
}

func (d *targetFileDAO) Delete(id string) error {
	_, err := d.db.Exec("DELETE FROM target_files WHERE id = ?", id)
	return err
}

type deployHistoryDAO struct {
	db *sql.DB
}

func NewDeployHistoryDAO(database *Database) DeployHistoryDAO {
	return &deployHistoryDAO{db: database.db}
}

func (d *deployHistoryDAO) GetAll(limit int) ([]DeployHistory, error) {
	query := `SELECT id, environment_id, environment_name, start_time, end_time, status, 
			  files, duration, error_message, created_at FROM deploy_history 
			  ORDER BY created_at DESC LIMIT ?`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []DeployHistory
	for rows.Next() {
		var h DeployHistory
		err := rows.Scan(&h.ID, &h.EnvironmentID, &h.EnvironmentName, &h.StartTime, &h.EndTime,
			&h.Status, &h.Files, &h.Duration, &h.ErrorMessage, &h.CreatedAt)
		if err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}
	return histories, nil
}

func (d *deployHistoryDAO) GetByEnvironmentID(envID string, limit int) ([]DeployHistory, error) {
	query := `SELECT id, environment_id, environment_name, start_time, end_time, status, 
			  files, duration, error_message, created_at FROM deploy_history 
			  WHERE environment_id = ? ORDER BY created_at DESC LIMIT ?`

	rows, err := d.db.Query(query, envID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []DeployHistory
	for rows.Next() {
		var h DeployHistory
		err := rows.Scan(&h.ID, &h.EnvironmentID, &h.EnvironmentName, &h.StartTime, &h.EndTime,
			&h.Status, &h.Files, &h.Duration, &h.ErrorMessage, &h.CreatedAt)
		if err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}
	return histories, nil
}

func (d *deployHistoryDAO) GetByID(id string) (*DeployHistory, error) {
	query := `SELECT id, environment_id, environment_name, start_time, end_time, status, 
			  files, duration, error_message, created_at FROM deploy_history WHERE id = ?`

	var h DeployHistory
	err := d.db.QueryRow(query, id).Scan(&h.ID, &h.EnvironmentID, &h.EnvironmentName, &h.StartTime, &h.EndTime,
		&h.Status, &h.Files, &h.Duration, &h.ErrorMessage, &h.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (d *deployHistoryDAO) Create(history *DeployHistory) error {
	query := `INSERT INTO deploy_history (id, environment_id, environment_name, start_time, end_time, 
			  status, files, duration, error_message, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, history.ID, history.EnvironmentID, history.EnvironmentName,
		history.StartTime, history.EndTime, history.Status, history.Files,
		history.Duration, history.ErrorMessage, history.CreatedAt)
	return err
}

func (d *deployHistoryDAO) Update(history *DeployHistory) error {
	query := `UPDATE deploy_history SET end_time = ?, status = ?, files = ?, 
			  duration = ?, error_message = ? WHERE id = ?`

	_, err := d.db.Exec(query, history.EndTime, history.Status, history.Files,
		history.Duration, history.ErrorMessage, history.ID)
	return err
}

func (d *deployHistoryDAO) Delete(id string) error {
	_, err := d.db.Exec("DELETE FROM deploy_history WHERE id = ?", id)
	return err
}

func (d *deployHistoryDAO) DeleteOld(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	_, err := d.db.Exec("DELETE FROM deploy_history WHERE created_at < ?", cutoff)
	return err
}

type deployLogDAO struct {
	db *sql.DB
}

func NewDeployLogDAO(database *Database) DeployLogDAO {
	return &deployLogDAO{db: database.db}
}

func (d *deployLogDAO) GetByDeployID(deployID string) ([]DeployLog, error) {
	query := `SELECT id, deploy_id, level, message, timestamp, created_at 
			  FROM deploy_logs WHERE deploy_id = ? ORDER BY timestamp ASC`

	rows, err := d.db.Query(query, deployID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []DeployLog
	for rows.Next() {
		var log DeployLog
		err := rows.Scan(&log.ID, &log.DeployID, &log.Level, &log.Message, &log.Timestamp, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (d *deployLogDAO) Create(log *DeployLog) error {
	query := `INSERT INTO deploy_logs (id, deploy_id, level, message, timestamp, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, log.ID, log.DeployID, log.Level, log.Message, log.Timestamp, log.CreatedAt)
	return err
}

func (d *deployLogDAO) BatchCreate(logs []DeployLog) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO deploy_logs (id, deploy_id, level, message, timestamp, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, log := range logs {
		if _, err := stmt.Exec(log.ID, log.DeployID, log.Level, log.Message, log.Timestamp, log.CreatedAt); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *deployLogDAO) DeleteByDeployID(deployID string) error {
	_, err := d.db.Exec("DELETE FROM deploy_logs WHERE deploy_id = ?", deployID)
	return err
}

type frontendDeploymentDAO struct {
	db *sql.DB
}

func NewFrontendDeploymentDAO(database *Database) FrontendDeploymentDAO {
	return &frontendDeploymentDAO{db: database.db}
}

func (d *frontendDeploymentDAO) GetAll() ([]FrontendDeployment, error) {
	query := `SELECT id, name, project_root, build_command, build_args, output_dir, 
			  environment_id, enabled, created_at, updated_at FROM frontend_deployments 
			  ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []FrontendDeployment
	for rows.Next() {
		var dep FrontendDeployment
		err := rows.Scan(&dep.ID, &dep.Name, &dep.ProjectRoot, &dep.BuildCommand,
			&dep.BuildArgs, &dep.OutputDir, &dep.EnvironmentID, &dep.Enabled,
			&dep.CreatedAt, &dep.UpdatedAt)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, dep)
	}
	return deployments, nil
}

func (d *frontendDeploymentDAO) GetByID(id string) (*FrontendDeployment, error) {
	query := `SELECT id, name, project_root, build_command, build_args, output_dir, 
			  environment_id, enabled, created_at, updated_at FROM frontend_deployments WHERE id = ?`

	var dep FrontendDeployment
	err := d.db.QueryRow(query, id).Scan(&dep.ID, &dep.Name, &dep.ProjectRoot, &dep.BuildCommand,
		&dep.BuildArgs, &dep.OutputDir, &dep.EnvironmentID, &dep.Enabled, &dep.CreatedAt, &dep.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

func (d *frontendDeploymentDAO) GetByEnvironmentID(envID string) ([]FrontendDeployment, error) {
	query := `SELECT id, name, project_root, build_command, build_args, output_dir, 
			  environment_id, enabled, created_at, updated_at FROM frontend_deployments 
			  WHERE environment_id = ? ORDER BY created_at DESC`

	rows, err := d.db.Query(query, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []FrontendDeployment
	for rows.Next() {
		var dep FrontendDeployment
		err := rows.Scan(&dep.ID, &dep.Name, &dep.ProjectRoot, &dep.BuildCommand,
			&dep.BuildArgs, &dep.OutputDir, &dep.EnvironmentID, &dep.Enabled,
			&dep.CreatedAt, &dep.UpdatedAt)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, dep)
	}
	return deployments, nil
}

func (d *frontendDeploymentDAO) GetEnabled() ([]FrontendDeployment, error) {
	query := `SELECT id, name, project_root, build_command, build_args, output_dir, 
			  environment_id, enabled, created_at, updated_at FROM frontend_deployments 
			  WHERE enabled = 1 ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []FrontendDeployment
	for rows.Next() {
		var dep FrontendDeployment
		err := rows.Scan(&dep.ID, &dep.Name, &dep.ProjectRoot, &dep.BuildCommand,
			&dep.BuildArgs, &dep.OutputDir, &dep.EnvironmentID, &dep.Enabled,
			&dep.CreatedAt, &dep.UpdatedAt)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, dep)
	}
	return deployments, nil
}

func (d *frontendDeploymentDAO) Create(deployment *FrontendDeployment) error {
	query := `INSERT INTO frontend_deployments (id, name, project_root, build_command, build_args, 
			  output_dir, environment_id, enabled, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, deployment.ID, deployment.Name, deployment.ProjectRoot,
		deployment.BuildCommand, deployment.BuildArgs, deployment.OutputDir,
		deployment.EnvironmentID, deployment.Enabled, now, now)
	return err
}

func (d *frontendDeploymentDAO) Update(deployment *FrontendDeployment) error {
	query := `UPDATE frontend_deployments SET name = ?, project_root = ?, build_command = ?, 
			  build_args = ?, output_dir = ?, environment_id = ?, enabled = ?, updated_at = ? WHERE id = ?`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, deployment.Name, deployment.ProjectRoot, deployment.BuildCommand,
		deployment.BuildArgs, deployment.OutputDir, deployment.EnvironmentID,
		deployment.Enabled, now, deployment.ID)
	return err
}

func (d *frontendDeploymentDAO) Delete(id string) error {
	_, err := d.db.Exec("DELETE FROM frontend_deployments WHERE id = ?", id)
	return err
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
