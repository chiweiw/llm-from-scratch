package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"deploy-tool/internal/logger"
	"deploy-tool/internal/model/entity"
	"deploy-tool/internal/utils"
)

type DeployService struct {
	progress         *entity.DeployProgress
	mavenBuild       *MavenBuildService
	cfgService       *ConfigService
	historyService   *HistoryService
	cancelFunc       context.CancelFunc
	currentHistoryID string
	wailsCtx         context.Context
}

func NewDeployService(cfgService *ConfigService, historyService *HistoryService, wailsCtx context.Context) *DeployService {
	return &DeployService{
		cfgService:     cfgService,
		historyService: historyService,
		wailsCtx:       wailsCtx,
		progress: &entity.DeployProgress{
			Status: entity.DeployStatusIdle,
		},
		mavenBuild: NewMavenBuildService(),
	}
}

func (s *DeployService) SetConfigService(cfg *ConfigService) {
	s.cfgService = cfg
}

func (s *DeployService) SetHistoryService(history *HistoryService) {
	s.historyService = history
}

func (s *DeployService) Start(envID string, jarIDs []string) error {
	if s.cfgService == nil {
		return fmt.Errorf("config service 未初始化")
	}

	env := s.cfgService.GetEnvironment(envID)
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
	s.currentHistoryID = historyID

	defer func() {
		s.currentHistoryID = ""
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
		logger.Info("开始逐台服务器上传并重启...")
		s.updateStepStatus("文件上传", entity.StepStatusRunning, "准备上传文件...")
		s.updateStepStatus("远程重启", entity.StepStatusRunning, "准备执行重启脚本...")
		if err := s.uploadAndRestartSequential(ctx, env, buildResult.BuiltFiles, jarIDs); err != nil {
			logger.Error("上传/重启失败: %v", err)
			// 哪个步骤失败就更新哪个步骤
			if strings.Contains(err.Error(), "上传") {
				s.updateStepStatus("文件上传", entity.StepStatusFailed, err.Error())
			} else {
				s.updateStepStatus("远程重启", entity.StepStatusFailed, err.Error())
			}
			s.updateProgressStatus(entity.DeployStatusFailed, err.Error())
			if s.historyService != nil && historyID != "" {
				s.historyService.Update(historyID, "failed", "", 0, err.Error())
			}
			return
		}
		s.updateStepStatus("文件上传", entity.StepStatusSuccess, "所有服务器上传完成")
		s.updateStepStatus("远程重启", entity.StepStatusSuccess, "所有服务器重启完成")
		logger.Info("逐台上传并重启完成")
		s.updateProgress(85)
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
	if s.cfgService == nil {
		return fmt.Errorf("config service 未初始化")
	}
	defaults := s.cfgService.GetSystemDefaults()
	result := CheckEnvironment(env, defaults)
	if result == nil {
		return fmt.Errorf("环境检查失败：未知错误")
	}
	if !result.Success {
		logger.Error("%s", result.Summary)
		for _, item := range result.Checks {
			switch item.Status {
			case entity.CheckStatusFail:
				if item.Message != "" {
					logger.Error("%s - %s", item.Name, item.Message)
				} else {
					logger.Error("%s", item.Name)
				}
			case entity.CheckStatusWarning:
				if item.Message != "" {
					logger.Warn("%s - %s", item.Name, item.Message)
				} else {
					logger.Warn("%s", item.Name)
				}
			default:
				if item.Message != "" {
					logger.Info("%s - %s", item.Name, item.Message)
				} else {
					logger.Info("%s", item.Name)
				}
			}
		}
		return fmt.Errorf("环境检查未通过")
	}
	// 仅有警告情况下，记录警告后继续
	for _, item := range result.Checks {
		if item.Status == entity.CheckStatusWarning {
			if item.Message != "" {
				logger.Warn("%s - %s", item.Name, item.Message)
			} else {
				logger.Warn("%s", item.Name)
			}
		}
	}
	return nil
}

func (s *DeployService) buildMavenConfig(env *entity.Environment) *MavenBuildConfig {
	defaults := s.cfgService.GetSystemDefaults()

	cfg := &MavenBuildConfig{
		ProjectRoot: env.ProjectRoot,
		MavenPath:   defaults.MavenPath,
		JavaHome:    defaults.JdkPath,
		Offline:     true,
		Quiet:       false,
		UseFPom:     true,
	}

	if defaults.MavenSettingsPath != "" {
		cfg.SettingsPath = defaults.MavenSettingsPath
	}

	if defaults.MavenRepoPath != "" {
		cfg.RepoLocal = defaults.MavenRepoPath
	}

	if len(defaults.MavenArgs) > 0 {
		// 过滤掉已由专用字段（SettingsPath / RepoLocal）处理的参数，避免重复追加
		var cleanGoals []string
		skipNext := false
		for _, arg := range defaults.MavenArgs {
			if skipNext {
				skipNext = false
				continue
			}
			if arg == "-s" || arg == "--settings" {
				skipNext = true // 跳过下一个参数（settings 文件路径）
				continue
			}
			if strings.HasPrefix(arg, "-Dmaven.repo.local=") {
				continue
			}
			cleanGoals = append(cleanGoals, arg)
		}
		cfg.Goals = cleanGoals
	}

	return cfg
}

func (s *DeployService) uploadFiles(ctx context.Context, env *entity.Environment, builtFiles []string, jarIDs []string) error {
	if len(env.Servers) == 0 {
		return fmt.Errorf("未配置任何服务器")
	}

	targetMap := make(map[string]entity.TargetFile)
	for _, t := range env.TargetFiles {
		targetMap[t.ID] = t
	}

	var selectedTargets []entity.TargetFile
	if len(jarIDs) == 0 {
		for _, t := range env.TargetFiles {
			if t.DefaultCheck {
				selectedTargets = append(selectedTargets, t)
			}
		}
	} else {
		for _, id := range jarIDs {
			if t, ok := targetMap[id]; ok {
				selectedTargets = append(selectedTargets, t)
			}
		}
	}

	if len(selectedTargets) == 0 {
		return fmt.Errorf("未选择任何要上传的文件")
	}

	logger.Info("准备上传 %d 个文件到 %d 台服务器", len(selectedTargets), len(env.Servers))

	for _, server := range env.Servers {
		logger.Info("连接服务器 %s (%s:%d)...", server.Name, server.Host, server.Port)

		client, err := utils.NewSFTPClient(server.Host, server.Port, server.Username, server.Password)
		if err != nil {
			return fmt.Errorf("连接服务器 %s 失败: %v", server.Name, err)
		}

		for _, t := range selectedTargets {
			localPath := t.LocalPath
			if !filepath.IsAbs(localPath) {
				localPath = filepath.Join(env.ProjectRoot, filepath.FromSlash(localPath))
			}

			remoteName := t.RemoteName
			if strings.TrimSpace(remoteName) == "" {
				remoteName = filepath.Base(localPath)
			}

			logger.Info("上传文件: %s -> %s:%s/%s", localPath, server.Host, server.DeployDir, remoteName)
			s.progress.CurrentFile = localPath
			s.progress.FileProgress = 0

			// 备份远程旧文件
			if err := client.BackupRemoteFile(server.DeployDir, remoteName, s.cfgService.GetSettings().BackupCleanup); err != nil {
				client.Close()
				return fmt.Errorf("服务器 %s 备份旧文件失败: %v", server.Name, err)
			}

			// 上传并上报进度
			start := time.Now()
			var lastReport time.Time
			var lastBytes int64
			onProgress := func(written, total int64) {
				percent := 0
				if total > 0 {
					percent = int((written * 100) / total)
				}
				s.progress.FileProgress = percent
				now := time.Now()
				if lastReport.IsZero() || now.Sub(lastReport) > 500*time.Millisecond {
					elapsed := now.Sub(start).Seconds()
					if elapsed > 0 {
						speed := float64(written) / 1024.0 / 1024.0 / elapsed
						s.progress.Speed = fmt.Sprintf("%.2f MB/s", speed)
					}
					lastReport = now
					lastBytes = written
				} else {
					_ = lastBytes
				}
			}

			if err := client.UploadFileWithProgress(localPath, server.DeployDir, remoteName, onProgress); err != nil {
				client.Close()
				return fmt.Errorf("上传文件到服务器 %s 失败: %v", server.Name, err)
			}

			s.progress.FileProgress = 100
		}

		client.Close()
	}

	return nil
}

func (s *DeployService) uploadAndRestartSequential(ctx context.Context, env *entity.Environment, builtFiles []string, jarIDs []string) error {
	if len(env.Servers) == 0 {
		return fmt.Errorf("未配置任何服务器")
	}

	targetMap := make(map[string]entity.TargetFile)
	for _, t := range env.TargetFiles {
		targetMap[t.ID] = t
	}
	var selectedTargets []entity.TargetFile
	if len(jarIDs) == 0 {
		for _, t := range env.TargetFiles {
			if t.DefaultCheck {
				selectedTargets = append(selectedTargets, t)
			}
		}
	} else {
		for _, id := range jarIDs {
			if t, ok := targetMap[id]; ok {
				selectedTargets = append(selectedTargets, t)
			}
		}
	}
	if len(selectedTargets) == 0 {
		return fmt.Errorf("未选择任何要上传的文件")
	}

	for idx, server := range env.Servers {
		stagePrefix := fmt.Sprintf("[%d/%d] 服务器 %s", idx+1, len(env.Servers), server.Name)
		logger.Info("%s: 连接中 (%s:%d)...", stagePrefix, server.Host, server.Port)
		client, err := utils.NewSFTPClient(server.Host, server.Port, server.Username, server.Password)
		if err != nil {
			return fmt.Errorf("%s: 连接失败: %v", stagePrefix, err)
		}

		// 上传前备份并上传所有目标
		for _, t := range selectedTargets {
			localPath := t.LocalPath
			if !filepath.IsAbs(localPath) {
				localPath = filepath.Join(env.ProjectRoot, filepath.FromSlash(localPath))
			}
			remoteName := t.RemoteName
			if strings.TrimSpace(remoteName) == "" {
				remoteName = filepath.Base(localPath)
			}
			logger.Info("%s: 备份远程旧文件 %s", stagePrefix, remoteName)
			if err := client.BackupRemoteFile(server.DeployDir, remoteName, s.cfgService.GetSettings().BackupCleanup); err != nil {
				client.Close()
				return fmt.Errorf("%s: 备份失败: %v", stagePrefix, err)
			}

			logger.Info("%s: 上传 %s -> %s:%s/%s", stagePrefix, localPath, server.Host, server.DeployDir, remoteName)
			s.progress.CurrentFile = localPath
			s.progress.FileProgress = 0
			start := time.Now()
			var lastReport time.Time
			onProgress := func(written, total int64) {
				percent := 0
				if total > 0 {
					percent = int((written * 100) / total)
				}
				s.progress.FileProgress = percent
				now := time.Now()
				if lastReport.IsZero() || now.Sub(lastReport) > 500*time.Millisecond {
					elapsed := now.Sub(start).Seconds()
					if elapsed > 0 {
						speed := float64(written) / 1024.0 / 1024.0 / elapsed
						s.progress.Speed = fmt.Sprintf("%.2f MB/s", speed)
					}
					lastReport = now
				}
			}
			if err := client.UploadFileWithProgress(localPath, server.DeployDir, remoteName, onProgress); err != nil {
				client.Close()
				return fmt.Errorf("%s: 上传失败: %v", stagePrefix, err)
			}
			s.progress.FileProgress = 100
		}

		// 重启该服务器
		if server.EnableRestart && strings.TrimSpace(server.RestartScript) != "" {
			s.updateStepStatus("远程重启", entity.StepStatusRunning, fmt.Sprintf("%s: 执行重启脚本", stagePrefix))
			cmd := server.RestartScript
			if server.UseSudo {
				cmd = "sudo " + cmd
			}
			output, err := client.RunCommand(cmd)
			if err != nil {
				client.Close()
				return fmt.Errorf("%s: 重启失败: %v，输出: %s", stagePrefix, err, strings.TrimSpace(output))
			}
			if strings.TrimSpace(output) != "" {
				logger.Info("%s: 重启输出: %s", stagePrefix, strings.TrimSpace(output))
			}
		} else {
			logger.Info("%s: 未启用重启或未配置脚本，跳过", stagePrefix)
		}

		client.Close()
	}

	return nil
}

func (s *DeployService) TryAddHistoryLog(level, message string) {
	if s.historyService == nil || s.currentHistoryID == "" {
		return
	}
	_ = s.historyService.AddLog(s.currentHistoryID, level, message)
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
