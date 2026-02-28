package services

import (
	"deploy-tool/internal/models"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckLocalConfig(env *models.Environment) []models.CheckItem {
	checks := []models.CheckItem{}

	if env.Local.ProjectRoot == "" {
		checks = append(checks, models.CheckItem{
			Name:    "项目根目录",
			Status:  models.CheckStatusFail,
			Message: "未配置项目根目录",
		})
		return checks
	}

	exists := checkPathExists(env.Local.ProjectRoot)
	if !exists {
		checks = append(checks, models.CheckItem{
			Name:    "项目根目录",
			Status:  models.CheckStatusFail,
			Message: "项目根目录不存在: " + env.Local.ProjectRoot,
		})
		return checks
	}

	checks = append(checks, models.CheckItem{
		Name:    "项目根目录",
		Status:  models.CheckStatusPass,
		Message: env.Local.ProjectRoot,
	})

	pomPath := filepath.Join(env.Local.ProjectRoot, "pom.xml")
	if !checkPathExists(pomPath) {
		checks = append(checks, models.CheckItem{
			Name:    "pom.xml 文件",
			Status:  models.CheckStatusFail,
			Message: "pom.xml 不存在于项目根目录",
		})
	}

	if env.Local.JdkPath == "" {
		checks = append(checks, models.CheckItem{
			Name:    "JDK 路径",
			Status:  models.CheckStatusWarning,
			Message: "未配置 JDK 路径（建议配置以避免问题）",
		})
	} else if !checkPathExists(env.Local.JdkPath) {
		checks = append(checks, models.CheckItem{
			Name:    "JDK 路径",
			Status:  models.CheckStatusWarning,
			Message: "JDK 路径不存在: " + env.Local.JdkPath,
		})
	} else {
		checks = append(checks, models.CheckItem{
			Name:    "JDK 路径",
			Status:  models.CheckStatusPass,
			Message: env.Local.JdkPath,
		})
	}

	if env.Local.MavenPath != "" {
		if !checkPathExists(env.Local.MavenPath) {
			checks = append(checks, models.CheckItem{
				Name:    "Maven 路径",
				Status:  models.CheckStatusWarning,
				Message: "Maven 路径不存在: " + env.Local.MavenPath,
			})
		} else {
			checks = append(checks, models.CheckItem{
				Name:    "Maven 路径",
				Status:  models.CheckStatusPass,
				Message: env.Local.MavenPath,
			})
		}
	} else {
		mavenPath := findMavenInPath()
		if mavenPath != "" {
			checks = append(checks, models.CheckItem{
				Name:    "Maven 路径",
				Status:  models.CheckStatusPass,
				Message: "已找到: " + mavenPath,
			})
		} else {
			checks = append(checks, models.CheckItem{
				Name:    "Maven 路径",
				Status:  models.CheckStatusWarning,
				Message: "未配置，将尝试使用系统 PATH 中的 mvn",
			})
		}
	}

	if env.Local.MavenSettingsPath != "" {
		if !checkPathExists(env.Local.MavenSettingsPath) {
			checks = append(checks, models.CheckItem{
				Name:    "Maven settings.xml",
				Status:  models.CheckStatusWarning,
				Message: "Maven settings.xml 不存在: " + env.Local.MavenSettingsPath,
			})
		} else {
			checks = append(checks, models.CheckItem{
				Name:    "Maven settings.xml",
				Status:  models.CheckStatusPass,
				Message: env.Local.MavenSettingsPath,
			})
		}
	}

	return checks
}

func findMavenInPath() string {
	cmd := exec.Command("mvn", "-version")
	cmd.Env = os.Environ()
	_, err := cmd.Output()
	if err != nil {
		return ""
	}
	mvnPath, err := exec.LookPath("mvn")
	if err != nil {
		return ""
	}
	return mvnPath
}

func CheckTargetFiles(env *models.Environment) []models.CheckItem {
	checks := []models.CheckItem{}

	if len(env.TargetFiles) == 0 {
		checks = append(checks, models.CheckItem{
			Name:    "目标文件配置",
			Status:  models.CheckStatusWarning,
			Message: "未配置目标文件，打包后需要手动确认",
		})
		return checks
	}

	checks = append(checks, models.CheckItem{
		Name:    "目标文件配置",
		Status:  models.CheckStatusPass,
		Message: fmt.Sprintf("已配置 %d 个目标文件", len(env.TargetFiles)),
	})

	for _, file := range env.TargetFiles {
		if file.LocalPath == "" {
			continue
		}
		fullPath := filepath.Join(env.Local.ProjectRoot, file.LocalPath)
		exists := checkPathExists(fullPath)
		if exists {
			checks = append(checks, models.CheckItem{
				Name:    "目标文件: " + file.LocalPath,
				Status:  models.CheckStatusPass,
				Message: "文件存在",
			})
		} else {
			checks = append(checks, models.CheckItem{
				Name:    "目标文件: " + file.LocalPath,
				Status:  models.CheckStatusWarning,
				Message: "文件不存在（需打包后生成）",
			})
		}
	}

	return checks
}

func CheckServers(env *models.Environment) []models.CheckItem {
	checks := []models.CheckItem{}

	if len(env.Servers) == 0 {
		checks = append(checks, models.CheckItem{
			Name:    "服务器配置",
			Status:  models.CheckStatusWarning,
			Message: "未配置服务器（仅本地打包）",
		})
		return checks
	}

	for _, server := range env.Servers {
		checkItem := CheckSFTPServer(&server)
		checks = append(checks, *checkItem)
	}

	return checks
}

func checkPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
