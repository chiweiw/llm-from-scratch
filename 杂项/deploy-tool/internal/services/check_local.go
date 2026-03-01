package services

import (
	"deploy-tool/internal/models"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CheckLocalConfig(env *models.Environment, defaults *models.SystemDefaultConfig) []models.CheckItem {
	checks := []models.CheckItem{}

	projectRoot := strings.TrimSpace(env.ProjectRoot)
	if projectRoot == "" {
		checks = append(checks, models.CheckItem{
			Name:    "项目根目录",
			Status:  models.CheckStatusFail,
			Message: "未配置项目根目录",
		})
		return checks
	}

	exists := checkPathExists(projectRoot)
	if !exists {
		checks = append(checks, models.CheckItem{
			Name:    "项目根目录",
			Status:  models.CheckStatusFail,
			Message: "项目根目录不存在: " + projectRoot,
		})
		return checks
	}

	checks = append(checks, models.CheckItem{
		Name:    "项目根目录",
		Status:  models.CheckStatusPass,
		Message: projectRoot,
	})

	pomPath := filepath.Join(projectRoot, "pom.xml")
	if !checkPathExists(pomPath) {
		checks = append(checks, models.CheckItem{
			Name:    "pom.xml 文件",
			Status:  models.CheckStatusFail,
			Message: "pom.xml 不存在于项目根目录",
		})
	}

	jdkPath := ""
	mavenPath := ""
	mavenSettingsPath := ""
	if defaults != nil {
		jdkPath = strings.TrimSpace(defaults.JdkPath)
		mavenPath = strings.TrimSpace(defaults.MavenPath)
		mavenSettingsPath = strings.TrimSpace(defaults.MavenSettingsPath)
	}

	if jdkPath == "" {
		checks = append(checks, models.CheckItem{
			Name:    "JDK 路径",
			Status:  models.CheckStatusWarning,
			Message: "未配置 JDK 路径（建议配置以避免问题）",
		})
	} else if !checkPathExists(jdkPath) {
		checks = append(checks, models.CheckItem{
			Name:    "JDK 路径",
			Status:  models.CheckStatusWarning,
			Message: "JDK 路径不存在: " + jdkPath,
		})
	} else {
		checks = append(checks, models.CheckItem{
			Name:    "JDK 路径",
			Status:  models.CheckStatusPass,
			Message: jdkPath,
		})
	}

	if mavenPath != "" {
		if !checkPathExists(mavenPath) {
			checks = append(checks, models.CheckItem{
				Name:    "Maven 路径",
				Status:  models.CheckStatusWarning,
				Message: "Maven 路径不存在: " + mavenPath,
			})
		} else {
			checks = append(checks, models.CheckItem{
				Name:    "Maven 路径",
				Status:  models.CheckStatusPass,
				Message: mavenPath,
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

	if mavenSettingsPath != "" {
		if !checkPathExists(mavenSettingsPath) {
			checks = append(checks, models.CheckItem{
				Name:    "Maven settings.xml",
				Status:  models.CheckStatusWarning,
				Message: "Maven settings.xml 不存在: " + mavenSettingsPath,
			})
		} else {
			checks = append(checks, models.CheckItem{
				Name:    "Maven settings.xml",
				Status:  models.CheckStatusPass,
				Message: mavenSettingsPath,
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
	projectRoot := strings.TrimSpace(env.ProjectRoot)
	if projectRoot == "" {
		return checks
	}

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
		fullPath := filepath.Join(projectRoot, file.LocalPath)
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
