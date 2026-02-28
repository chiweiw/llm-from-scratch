package services

import (
	"deploy-tool/internal/models"
	"fmt"
)

func CheckEnvironment(env *models.Environment) *models.CheckResult {
	result := &models.CheckResult{
		Success: true,
		Checks:  []models.CheckItem{},
		Summary: "",
	}

	localChecks := CheckLocalConfig(env)
	result.Checks = append(result.Checks, localChecks...)

	targetChecks := CheckTargetFiles(env)
	result.Checks = append(result.Checks, targetChecks...)

	serverChecks := CheckServers(env)
	result.Checks = append(result.Checks, serverChecks...)

	errorCount := 0
	warningCount := 0
	for _, check := range result.Checks {
		if check.Status == models.CheckStatusFail {
			errorCount++
		} else if check.Status == models.CheckStatusWarning {
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
