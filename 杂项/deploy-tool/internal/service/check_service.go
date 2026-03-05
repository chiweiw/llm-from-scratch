package service

import (
	"deploy-tool/internal/model/entity"
	"deploy-tool/internal/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CheckEnvironment(env *entity.Environment, defaults entity.SystemDefaultConfig) *entity.CheckResult {
	result := &entity.CheckResult{
		Success: true,
		Checks:  []entity.CheckItem{},
		Summary: "",
	}

	localChecks := CheckLocalConfig(env, defaults)
	result.Checks = append(result.Checks, localChecks...)

	targetChecks := CheckTargetFiles(env)
	result.Checks = append(result.Checks, targetChecks...)

	serverChecks := CheckServers(env)
	result.Checks = append(result.Checks, serverChecks...)

	errorCount := 0
	warningCount := 0
	for _, check := range result.Checks {
		if check.Status == entity.CheckStatusFail {
			errorCount++
		} else if check.Status == entity.CheckStatusWarning {
			warningCount++
		}
	}

	result.Success = errorCount == 0

	if errorCount > 0 {
		result.Summary = fmt.Sprintf("检查未通过: %d 个错误", errorCount)
		if warningCount > 0 {
			result.Summary += fmt.Sprintf(", %d 个警告", warningCount)
		}
	} else if warningCount > 0 {
		result.Summary = fmt.Sprintf("检查通过: %d 个警告", warningCount)
	} else {
		result.Summary = fmt.Sprintf("检查通过: %d 项检查全部正常", len(result.Checks))
	}

	return result
}

func CheckLocalConfig(env *entity.Environment, defaults entity.SystemDefaultConfig) []entity.CheckItem {
	checks := []entity.CheckItem{}

	projectRoot := strings.TrimSpace(env.ProjectRoot)
	if projectRoot == "" {
		checks = append(checks, entity.CheckItem{
			Name:    "项目根目录",
			Status:  entity.CheckStatusFail,
			Message: "未配置项目根目录",
		})
		return checks
	}

	exists := checkPathExists(projectRoot)
	if !exists {
		checks = append(checks, entity.CheckItem{
			Name:    "项目根目录",
			Status:  entity.CheckStatusFail,
			Message: "项目根目录不存在: " + projectRoot,
		})
		return checks
	}

	checks = append(checks, entity.CheckItem{
		Name:    "项目根目录",
		Status:  entity.CheckStatusPass,
		Message: projectRoot,
	})

	if strings.ToLower(strings.TrimSpace(env.BuildType)) == "frontend" {
		packagePath := filepath.Join(projectRoot, "package.json")
		if !checkPathExists(packagePath) {
			checks = append(checks, entity.CheckItem{
				Name:    "package.json 文件",
				Status:  entity.CheckStatusFail,
				Message: "package.json 不存在于项目根目录",
			})
		}
	} else {
		pomPath := filepath.Join(projectRoot, "pom.xml")
		if !checkPathExists(pomPath) {
			checks = append(checks, entity.CheckItem{
				Name:    "pom.xml 文件",
				Status:  entity.CheckStatusFail,
				Message: "pom.xml 不存在于项目根目录",
			})
		}

		mavenPath := strings.TrimSpace(defaults.MavenPath)
		mavenSettingsPath := strings.TrimSpace(defaults.MavenSettingsPath)

		if mavenPath != "" {
			if !checkPathExists(mavenPath) {
				checks = append(checks, entity.CheckItem{
					Name:    "Maven 路径",
					Status:  entity.CheckStatusWarning,
					Message: "Maven 路径不存在: " + mavenPath,
				})
			} else {
				checks = append(checks, entity.CheckItem{
					Name:    "Maven 路径",
					Status:  entity.CheckStatusPass,
					Message: mavenPath,
				})
			}
		} else {
			found := findMavenInPath()
			if found != "" {
				checks = append(checks, entity.CheckItem{
					Name:    "Maven 路径",
					Status:  entity.CheckStatusPass,
					Message: "已找到: " + found,
				})
			} else {
				checks = append(checks, entity.CheckItem{
					Name:    "Maven 路径",
					Status:  entity.CheckStatusWarning,
					Message: "未配置，将尝试使用系统 PATH 中的 mvn",
				})
			}
		}

		if mavenSettingsPath != "" {
			if !checkPathExists(mavenSettingsPath) {
				checks = append(checks, entity.CheckItem{
					Name:    "Maven settings.xml",
					Status:  entity.CheckStatusWarning,
					Message: "Maven settings.xml 不存在: " + mavenSettingsPath,
				})
			} else {
				checks = append(checks, entity.CheckItem{
					Name:    "Maven settings.xml",
					Status:  entity.CheckStatusPass,
					Message: mavenSettingsPath,
				})
			}
		}
	}

	return checks
}

func CheckTargetFiles(env *entity.Environment) []entity.CheckItem {
	checks := []entity.CheckItem{}
	projectRoot := strings.TrimSpace(env.ProjectRoot)
	if projectRoot == "" {
		return checks
	}

	if len(env.TargetFiles) == 0 {
		checks = append(checks, entity.CheckItem{
			Name:    "目标文件配置",
			Status:  entity.CheckStatusWarning,
			Message: "未配置目标文件，打包后需要手动确认",
		})
		return checks
	}

	checks = append(checks, entity.CheckItem{
		Name:    "目标文件配置",
		Status:  entity.CheckStatusPass,
		Message: fmt.Sprintf("已配置 %d 个目标文件", len(env.TargetFiles)),
	})

	for _, file := range env.TargetFiles {
		if file.LocalPath == "" {
			continue
		}
		if strings.ToLower(strings.TrimSpace(env.BuildType)) == "frontend" {
			urlPath := strings.TrimSpace(file.UrlPath)
			if urlPath == "" {
				checks = append(checks, entity.CheckItem{
					Name:    "URL 路径: " + file.LocalPath,
					Status:  entity.CheckStatusFail,
					Message: "未配置 URL 路径",
				})
			} else if !strings.HasPrefix(urlPath, "/") {
				checks = append(checks, entity.CheckItem{
					Name:    "URL 路径: " + file.LocalPath,
					Status:  entity.CheckStatusFail,
					Message: "URL 路径必须以 / 开头",
				})
			} else {
				checks = append(checks, entity.CheckItem{
					Name:    "URL 路径: " + file.LocalPath,
					Status:  entity.CheckStatusPass,
					Message: urlPath,
				})
			}
			continue
		}
		fullPath := filepath.Join(projectRoot, file.LocalPath)
		exists := checkPathExists(fullPath)
		if exists {
			checks = append(checks, entity.CheckItem{
				Name:    "目标文件: " + file.LocalPath,
				Status:  entity.CheckStatusPass,
				Message: "文件存在",
			})
		} else {
			checks = append(checks, entity.CheckItem{
				Name:    "目标文件: " + file.LocalPath,
				Status:  entity.CheckStatusWarning,
				Message: "文件不存在（需打包后生成）",
			})
		}
	}

	return checks
}

func CheckServers(env *entity.Environment) []entity.CheckItem {
	checks := []entity.CheckItem{}

	if len(env.Servers) == 0 {
		checks = append(checks, entity.CheckItem{
			Name:    "服务器配置",
			Status:  entity.CheckStatusWarning,
			Message: "未配置服务器（仅本地打包）",
		})
		return checks
	}

	for _, server := range env.Servers {
		checkItem := utils.CheckSFTPServer(&server)
		checks = append(checks, *checkItem)
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

func checkPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
