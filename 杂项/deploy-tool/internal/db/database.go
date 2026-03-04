package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

var (
	instance *Database
)

func Init(dbPath string) (*Database, error) {
	if dbPath == "" {
		exePath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("获取可执行文件路径失败: %w", err)
		}
		dir := filepath.Dir(exePath)
		dbPath = filepath.Join(dir, "deploy-tool.db")
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	database := &Database{db: db}

	if err := database.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	if err := database.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("运行迁移失败: %w", err)
	}

	instance = database
	return database, nil
}

func GetInstance() *Database {
	return instance
}

func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *Database) createTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS global_settings (
			id TEXT PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
		)`,
		`CREATE TABLE IF NOT EXISTS system_defaults (
			id TEXT PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
		)`,
		`CREATE TABLE IF NOT EXISTS environments (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			identifier TEXT NOT NULL,
			description TEXT,
			project_root TEXT NOT NULL,
			cloud_deploy INTEGER NOT NULL DEFAULT 0,
			timeout INTEGER NOT NULL DEFAULT 600,
			backup_cleanup INTEGER NOT NULL DEFAULT 0,
			check_status TEXT DEFAULT 'unchecked',
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
		)`,
		`CREATE TABLE IF NOT EXISTS server_configs (
			id TEXT PRIMARY KEY,
			environment_id TEXT NOT NULL,
			name TEXT NOT NULL,
			host TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 22,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			deploy_dir TEXT NOT NULL,
			restart_script TEXT,
			enable_restart INTEGER NOT NULL DEFAULT 0,
			use_sudo INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS target_files (
			id TEXT PRIMARY KEY,
			environment_id TEXT NOT NULL,
			local_path TEXT NOT NULL,
			remote_name TEXT,
			default_check INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS deploy_history (
			id TEXT PRIMARY KEY,
			environment_id TEXT NOT NULL,
			environment_name TEXT NOT NULL,
			start_time INTEGER NOT NULL,
			end_time INTEGER,
			status TEXT NOT NULL,
			files TEXT,
			duration INTEGER,
			error_message TEXT,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS frontend_deployments (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			project_root TEXT NOT NULL,
			build_command TEXT NOT NULL,
			build_args TEXT,
			output_dir TEXT NOT NULL,
			environment_id TEXT,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS deploy_logs (
			id TEXT PRIMARY KEY,
			deploy_id TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			FOREIGN KEY (deploy_id) REFERENCES deploy_history(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
		)`,
	}

	for _, sql := range tables {
		if _, err := d.db.Exec(sql); err != nil {
			return fmt.Errorf("创建表失败: %w", err)
		}
	}

	return nil
}

func (d *Database) runMigrations() error {
	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "001_initial_schema",
			sql:  "",
		},
		{
			name: "002_add_indexes",
			sql: `CREATE INDEX IF NOT EXISTS idx_environments_identifier ON environments(identifier);
				CREATE INDEX IF NOT EXISTS idx_server_configs_env ON server_configs(environment_id);
				CREATE INDEX IF NOT EXISTS idx_target_files_env ON target_files(environment_id);
				CREATE INDEX IF NOT EXISTS idx_deploy_history_env ON deploy_history(environment_id);
				CREATE INDEX IF NOT EXISTS idx_deploy_logs_deploy ON deploy_logs(deploy_id);
				CREATE INDEX IF NOT EXISTS idx_frontend_deployments_env ON frontend_deployments(environment_id);`,
		},
		{
			name: "003_add_frontend_support",
			sql:  "",
		},
	}

	for _, migration := range migrations {
		applied, err := d.isMigrationApplied(migration.name)
		if err != nil {
			return err
		}

		if applied {
			continue
		}

		if migration.sql != "" {
			if _, err := d.db.Exec(migration.sql); err != nil {
				return fmt.Errorf("执行迁移 %s 失败: %w", migration.name, err)
			}
		}

		if _, err := d.db.Exec(
			"INSERT INTO migrations (name, applied_at) VALUES (?, ?)",
			migration.name, time.Now().Unix(),
		); err != nil {
			return fmt.Errorf("记录迁移 %s 失败: %w", migration.name, err)
		}
	}

	return nil
}

func (d *Database) isMigrationApplied(name string) (bool, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *Database) BeginTx() (*sql.Tx, error) {
	return d.db.Begin()
}
