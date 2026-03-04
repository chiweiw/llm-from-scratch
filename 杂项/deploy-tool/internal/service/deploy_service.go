package service

import (
	"context"
	"fmt"
	"time"

	"deploy-tool/internal/logger"
	"deploy-tool/internal/model/entity"
)

type DeployService struct {
	progress       *entity.DeployProgress
	mavenBuild     *MavenBuildService
	configService  *ConfigService
	historyService *HistoryService
	cancelFunc     context.CancelFunc
}

func NewDeployService() *DeployService {
	return &DeployService{
		progress: &entity.DeployProgress{
			Status: entity.DeployStatusIdle,
		},
		mavenBuild: NewMavenBuildService(),
	}
}

func (s *DeployService) SetConfigService(cfg *ConfigService) {
	s.configService = cfg
}

func (s *DeployService) SetHistoryService(history *HistoryService) {
	s.historyService = history
}

func (s *DeployService) Start(envID string, jarIDs []string) error {
	if s.configService == nil {
		return fmt.Errorf("config service 未初始化")
	}

	env := s.configService.GetEnvironment(envID)
	if env == nil {
		return fmt.Errorf("环境不存在: %s", envID)
	}

	s.progress = &entity.DeployProgress{
		EnvironmentID: envID,
		Status:        entity.DeployStatusRunning,
		StartTime:     time.Now().Unix(),
		TotalProgress: 0,
		Steps: []entity.StepProgress{
			{Name: "环境检查", Status: entity.StepStatusPending, Progress: 0},
			{Name: "Maven 打包", Status: entity.StepStatusPending, Progress: 0},
			{Name: "文件上传", Status: entity.StepStatusPending, Progress: 0},
			{Name: "远程重启", Status: entity.StepStatusPending, Progress: 0},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	logger.Info("准备开始部署，环境: %s", env.Name)

	go s.executeDeploy(ctx, env, jarIDs)

	logger.Info("部署任务已启动，环境: %s, Jar IDs: %v", envID, jarIDs)

	return nil
}

func (s *DeployService) executeDeploy(ctx context.Context, env *entity.Environment, jarIDs []string) {
	var historyID string
	if s.historyService != nil {
		historyID = s.historyService.Create(env.ID, env.Name)
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("部署任务异常: %v", r)
			s.updateProgressStatus(entity.DeployStatusFailed, fmt.Sprintf("部署失败: %v", r))
			if s.historyService != nil && historyID != "" {
				s.historyService.Update(historyID, "failed", "", 0, fmt.Sprintf("部署失败: %v", r))
			}
		}
	}()

	logger.Info("开始执行部署流程，环境: %s, 项目根目录: %s", env.Name, env.ProjectRoot)

	s.updateStepStatus("环境检查", entity.StepStatusRunning, "正在检查环境配置...")
	logger.Info("开始环境检查...")
	if err := s.checkEnvironment(env); err != nil {
		logger.Error("环境检查失败: %v", err)
		s.updateStepStatus("环境检查", entity.StepStatusFailed, err.Error())
		s.updateProgressStatus(entity.DeployStatusFailed, err.Error())
		if s.historyService != nil && historyID != "" {
			s.historyService.Update(historyID, "failed", "", 0, err.Error())
		}
		return
	}
	s.updateStepStatus("环境检查", entity.StepStatusSuccess, "环境检查通过")
	logger.Info("环境检查通过")
	s.updateProgress(10)

	logger.Info("开始 Maven 打包...")
	s.updateStepStatus("Maven 打包", entity.StepStatusRunning, "准备 Maven 构建...")
	mavenCfg := s.buildMavenConfig(env)
	buildResult, err := s.mavenBuild.StartBuild(ctx, mavenCfg, s.progress)
	if err != nil {
		logger.Error("Maven 打包失败: %v", err)
		s.updateProgressStatus(entity.DeployStatusFailed, fmt.Sprintf("Maven 打包失败: %v", err))
		if s.historyService != nil && historyID != "" {
			s.historyService.Update(historyID, "failed", "", 0, fmt.Sprintf("Maven 打包失败: %v", err))
		}
		return
	}

	if !buildResult.Success {
		logger.Error("Maven 打包失败: %s", buildResult.ErrorMessage)
		s.updateProgressStatus(entity.DeployStatusFailed, buildResult.ErrorMessage)
		if s.historyService != nil && historyID != "" {
			s.historyService.Update(historyID, "failed", "", 0, buildResult.ErrorMessage)
		}
		return
	}

	s.updateProgress(50)
	logger.Info("Maven 打包完成，构建产物: %v", buildResult.BuiltFiles)

	if env.CloudDeploy {
		logger.Info("开始上传文件到服务器...")
		s.updateStepStatus("文件上传", entity.StepStatusRunning, "准备上传文件...")
		if err := s.uploadFiles(ctx, env, buildResult.BuiltFiles, jarIDs); err != nil {
			logger.Error("文件上传失败: %v", err)
			s.updateStepStatus("文件上传", entity.StepStatusFailed, err.Error())
			s.updateProgressStatus(entity.DeployStatusFailed, err.Error())
			if s.historyService != nil && historyID != "" {
				s.historyService.Update(historyID, "failed", "", 0, err.Error())
			}
			return
		}
		s.updateStepStatus("文件上传", entity.StepStatusSuccess, "文件上传完成")
		logger.Info("文件上传完成")
		s.updateProgress(80)

		logger.Info("开始远程重启服务...")
		s.updateStepStatus("远程重启", entity.StepStatusRunning, "准备执行重启脚本...")
		if err := s.restartServices(ctx, env); err != nil {
			logger.Error("远程重启失败: %v", err)
			s.updateStepStatus("远程重启", entity.StepStatusFailed, err.Error())
			s.updateProgressStatus(entity.DeployStatusFailed, err.Error())
			if s.historyService != nil && historyID != "" {
				s.historyService.Update(historyID, "failed", "", 0, err.Error())
			}
			return
		}
		s.updateStepStatus("远程重启", entity.StepStatusSuccess, "服务重启完成")
		logger.Info("远程重启完成")
	} else {
		logger.Info("云端部署未启用，跳过文件上传和远程重启")
		s.updateStepStatus("文件上传", entity.StepStatusSkipped, "云端部署未启用")
		s.updateStepStatus("远程重启", entity.StepStatusSkipped, "云端部署未启用")
	}

	s.updateProgress(100)
	s.updateProgressStatus(entity.DeployStatusSuccess, "部署完成")
	logger.Info("部署流程完成")

	if s.historyService != nil && historyID != "" {
		duration := time.Now().Unix() - s.progress.StartTime
		files := ""
		if len(buildResult.BuiltFiles) > 0 {
			for _, f := range buildResult.BuiltFiles {
				if files != "" {
					files += ", "
				}
				files += f
			}
		}
		s.historyService.Update(historyID, "success", files, duration, "")
	}
}

func (s *DeployService) checkEnvironment(env *entity.Environment) error {
	if s.configService == nil {
		return fmt.Errorf("config service 未初始化")
	}
	defaults := s.configService.GetSystemDefaults()
	result := CheckEnvironment(env, defaults)
	if result == nil {
		return fmt.Errorf("环境检查失败：未知错误")
	}
	if !result.Success {
		logger.Error("%s", result.Summary)
		for _, item := range result.Checks {
			switch item.Status {
			case entity.CheckStatusFail:
				if item.Message != "" { logger.Error("%s - %s", item.Name, item.Message) } else { logger.Error("%s", item.Name) }
			case entity.CheckStatusWarning:
				if item.Message != "" { logger.Warn("%s - %s", item.Name, item.Message) } else { logger.Warn("%s", item.Name) }
			default:
				if item.Message != "" { logger.Info("%s - %s", item.Name, item.Message) } else { logger.Info("%s", item.Name) }
			}
		}
		return fmt.Errorf("环境检查未通过")
	}
	// 仅有警告情况下，记录警告后继续
	for _, item := range result.Checks {
		if item.Status == entity.CheckStatusWarning {
			if item.Message != "" { logger.Warn("%s - %s", item.Name, item.Message) } else { logger.Warn("%s", item.Name) }
		}
	}
	return nil
}

func (s *DeployService) buildMavenConfig(env *entity.Environment) *MavenBuildConfig {
	defaults := s.configService.GetSystemDefaults()

	cfg := &MavenBuildConfig{
		ProjectRoot: env.ProjectRoot,
		MavenPath:   defaults.MavenPath,
		JavaHome:    defaults.JdkPath,
		Offline:     true,
		Quiet:       true,
		UseFPom:     true,
	}

	if defaults.MavenSettingsPath != "" {
		cfg.SettingsPath = defaults.MavenSettingsPath
	}

	if defaults.MavenRepoPath != "" {
		cfg.RepoLocal = defaults.MavenRepoPath
	}

	if len(defaults.MavenArgs) > 0 {
		cfg.Goals = defaults.MavenArgs
	}

	return cfg
}

func (s *DeployService) uploadFiles(ctx context.Context, env *entity.Environment, builtFiles []string, jarIDs []string) error {
	return fmt.Errorf("文件上传功能待实现")
}

func (s *DeployService) restartServices(ctx context.Context, env *entity.Environment) error {
	return fmt.Errorf("远程重启功能待实现")
}

func (s *DeployService) Cancel() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	s.mavenBuild.Cancel()
	if s.progress != nil {
		s.progress.Status = entity.DeployStatusCanceled
	}
}

func (s *DeployService) GetProgress() *entity.DeployProgress {
	return s.progress
}

func (s *DeployService) updateProgressStatus(status string, message string) {
	if s.progress == nil {
		return
	}
	s.progress.Status = status
	s.progress.ErrorMessage = message
}

func (s *DeployService) updateProgress(percent int) {
	if s.progress == nil {
		return
	}
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}
	s.progress.TotalProgress = percent
}

func (s *DeployService) updateStepStatus(stepName string, status string, message string) {
	if s.progress == nil {
		return
	}

	found := false
	for i, step := range s.progress.Steps {
		if step.Name == stepName {
			s.progress.Steps[i].Status = status
			s.progress.Steps[i].Message = message
			found = true
			break
		}
	}

	if !found {
		s.progress.Steps = append(s.progress.Steps, entity.StepProgress{
			Name:    stepName,
			Status:  status,
			Message: message,
		})
	}

	s.progress.CurrentStep = stepName
}
