package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	if strings.TrimSpace(env.BuildType) == "" {
		env.BuildType = "backend"
	}
	buildType := strings.ToLower(strings.TrimSpace(env.BuildType))
	var steps []entity.StepProgress
	if buildType == "frontend" {
		steps = []entity.StepProgress{
			{Name: "环境检查", Status: entity.StepStatusPending, Progress: 0},
			{Name: "前端打包", Status: entity.StepStatusPending, Progress: 0},
			{Name: "压缩 dist", Status: entity.StepStatusPending, Progress: 0},
			{Name: "文件上传", Status: entity.StepStatusPending, Progress: 0},
			{Name: "远程备份", Status: entity.StepStatusPending, Progress: 0},
			{Name: "远程解压", Status: entity.StepStatusPending, Progress: 0},
		}
	} else {
		steps = []entity.StepProgress{
			{Name: "环境检查", Status: entity.StepStatusPending, Progress: 0},
			{Name: "Maven 打包", Status: entity.StepStatusPending, Progress: 0},
			{Name: "文件上传", Status: entity.StepStatusPending, Progress: 0},
			{Name: "远程重启", Status: entity.StepStatusPending, Progress: 0},
		}
	}

	s.progress = &entity.DeployProgress{
		EnvironmentID: envID,
		Status:        entity.DeployStatusRunning,
		StartTime:     time.Now().Unix(),
		TotalProgress: 0,
		Steps:         steps,
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

	if strings.ToLower(strings.TrimSpace(env.BuildType)) == "frontend" {
		files, err := s.deployFrontend(ctx, env)
		if err != nil {
			logger.Error("前端部署失败: %v", err)
			s.updateProgressStatus(entity.DeployStatusFailed, err.Error())
			if s.historyService != nil && historyID != "" {
				s.historyService.Update(historyID, "failed", "", 0, err.Error())
			}
			return
		}
		s.updateProgress(100)
		s.updateProgressStatus(entity.DeployStatusSuccess, "部署完成")
		logger.Info("部署流程完成")
		if s.historyService != nil && historyID != "" {
			duration := time.Now().Unix() - s.progress.StartTime
			s.historyService.Update(historyID, "success", files, duration, "")
		}
		return
	}

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
	settings := s.cfgService.GetSettings()

	cfg := &MavenBuildConfig{
		ProjectRoot: env.ProjectRoot,
		MavenPath:   defaults.MavenPath,
		JavaHome:    defaults.JdkPath,
		Offline:     settings.OfflineBuild,
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
		if ctx.Err() != nil {
			return ctx.Err()
		}
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

func (s *DeployService) deployFrontend(ctx context.Context, env *entity.Environment) (string, error) {
	s.updateStepStatus("前端打包", entity.StepStatusRunning, "执行 npm run build")
	if err := s.runNpmBuild(ctx, env.ProjectRoot); err != nil {
		s.updateStepStatus("前端打包", entity.StepStatusFailed, err.Error())
		return "", fmt.Errorf("前端打包失败: %v", err)
	}
	s.updateStepStatus("前端打包", entity.StepStatusSuccess, "打包完成")
	s.updateProgress(40)

	s.updateStepStatus("压缩 dist", entity.StepStatusRunning, "正在压缩 dist...")
	zipPath, err := s.zipFrontendDist(env.ProjectRoot)
	if err != nil {
		s.updateStepStatus("压缩 dist", entity.StepStatusFailed, err.Error())
		return "", fmt.Errorf("压缩 dist 失败: %v", err)
	}
	s.updateStepStatus("压缩 dist", entity.StepStatusSuccess, filepath.Base(zipPath))
	s.updateProgress(50)

	target := pickFrontendTarget(env)

	if !env.CloudDeploy {
		s.updateStepStatus("文件上传", entity.StepStatusSkipped, "云端部署未启用")
		s.updateStepStatus("远程备份", entity.StepStatusSkipped, "云端部署未启用")
		s.updateStepStatus("远程解压", entity.StepStatusSkipped, "云端部署未启用")
		return zipPath, nil
	}

	s.updateStepStatus("文件上传", entity.StepStatusRunning, "准备上传 dist.zip")
	s.updateStepStatus("远程备份", entity.StepStatusRunning, "准备备份 dist")
	s.updateStepStatus("远程解压", entity.StepStatusRunning, "准备解压 dist.zip")

	if err := s.deployFrontendToServers(ctx, env, zipPath, target); err != nil {
		if strings.Contains(err.Error(), "上传") {
			s.updateStepStatus("文件上传", entity.StepStatusFailed, err.Error())
		} else if strings.Contains(err.Error(), "备份") {
			s.updateStepStatus("远程备份", entity.StepStatusFailed, err.Error())
		} else if strings.Contains(err.Error(), "解压") {
			s.updateStepStatus("远程解压", entity.StepStatusFailed, err.Error())
		} else {
			s.updateStepStatus("文件上传", entity.StepStatusFailed, err.Error())
		}
		return "", err
	}

	s.updateStepStatus("文件上传", entity.StepStatusSuccess, "所有服务器上传完成")
	s.updateStepStatus("远程备份", entity.StepStatusSuccess, "所有服务器备份完成")
	s.updateStepStatus("远程解压", entity.StepStatusSuccess, "所有服务器解压完成")
	s.updateProgress(90)

	return zipPath, nil
}

func (s *DeployService) runNpmBuild(ctx context.Context, projectRoot string) error {
	if projectRoot == "" {
		return fmt.Errorf("项目根目录不能为空")
	}
	if _, err := os.Stat(projectRoot); os.IsNotExist(err) {
		return fmt.Errorf("项目根目录不存在: %s", projectRoot)
	}

	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("未找到 npm 可执行文件")
	}
	if _, err := exec.Command("npm", "-v").Output(); err != nil {
		return fmt.Errorf("npm 校验失败: %v", err)
	}

	cmd := exec.CommandContext(ctx, "npm", "run", "build")
	cmd.Dir = projectRoot

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取 stdout 失败: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("获取 stderr 失败: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 npm build 失败: %v", err)
	}

	stream := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				logger.Info("%s", line)
			}
		}
	}
	go stream(stdout)
	go stream(stderr)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("npm build 失败: %v", err)
	}
	return nil
}

func (s *DeployService) zipFrontendDist(projectRoot string) (string, error) {
	distDir := filepath.Join(projectRoot, "dist")
	info, err := os.Stat(distDir)
	if err != nil {
		return "", fmt.Errorf("dist 目录不存在: %v", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("dist 不是目录")
	}

	zipPath := filepath.Join(projectRoot, "dist.zip")
	if _, err := os.Stat(zipPath); err == nil {
		_ = os.Remove(zipPath)
	}
	if err := utils.ZipDir(distDir, zipPath); err != nil {
		return "", err
	}
	return zipPath, nil
}

func (s *DeployService) deployFrontendToServers(ctx context.Context, env *entity.Environment, zipPath string, target entity.TargetFile) error {
	if len(env.Servers) == 0 {
		return fmt.Errorf("未配置任何服务器")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	remoteName := "dist.zip"
	if strings.TrimSpace(target.RemoteName) != "" {
		remoteName = target.RemoteName
	}
	baseDir := normalizeUrlPath(target.UrlPath)
	for idx, server := range env.Servers {
		stagePrefix := fmt.Sprintf("[%d/%d] 服务器 %s", idx+1, len(env.Servers), server.Name)
		logger.Info("%s: 连接中 (%s:%d)...", stagePrefix, server.Host, server.Port)
		client, err := utils.NewSFTPClient(server.Host, server.Port, server.Username, server.Password)
		if err != nil {
			return fmt.Errorf("%s: 连接失败: %v", stagePrefix, err)
		}

		logger.Info("%s: 上传 dist.zip -> %s:%s/%s", stagePrefix, server.Host, baseDir, remoteName)
		s.progress.CurrentFile = zipPath
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
		if err := client.UploadFileWithProgress(zipPath, baseDir, remoteName, onProgress); err != nil {
			client.Close()
			return fmt.Errorf("%s: 上传失败: %v", stagePrefix, err)
		}
		s.progress.FileProgress = 100

		ts := time.Now().Format("20060102150405")
		backupCmd := fmt.Sprintf("cd %s && if [ -d dist ]; then mv dist dist.%s.bak; fi", shellQuote(baseDir), ts)
		backupCmd = wrapSudo(backupCmd, server.UseSudo)
		if output, err := client.RunCommand(backupCmd); err != nil {
			client.Close()
			return fmt.Errorf("%s: 备份失败: %v，输出: %s", stagePrefix, err, strings.TrimSpace(output))
		}

		tmpDir := "__dist_tmp__" + ts
		unzipCmd := fmt.Sprintf("cd %s && rm -rf %s && rm -rf dist && (unzip -o %s -d %s >/dev/null 2>&1 || python3 -m zipfile -e %s %s) && mv %s/dist dist && rm -rf %s && rm -f %s",
			shellQuote(baseDir),
			shellQuote(tmpDir),
			shellQuote(remoteName),
			shellQuote(tmpDir),
			shellQuote(remoteName),
			shellQuote(tmpDir),
			shellQuote(tmpDir),
			shellQuote(tmpDir),
			shellQuote(remoteName),
		)
		unzipCmd = wrapSudo(unzipCmd, server.UseSudo)
		output, err := client.RunCommand(unzipCmd)
		if err != nil {
			client.Close()
			return fmt.Errorf("%s: 解压失败: %v，输出: %s", stagePrefix, err, strings.TrimSpace(output))
		}
		if strings.TrimSpace(output) != "" {
			logger.Info("%s: 解压输出: %s", stagePrefix, strings.TrimSpace(output))
		}

		client.Close()
	}
	return nil
}

func pickFrontendTarget(env *entity.Environment) entity.TargetFile {
	for _, t := range env.TargetFiles {
		if t.DefaultCheck {
			return t
		}
	}
	if len(env.TargetFiles) > 0 {
		return env.TargetFiles[0]
	}
	return entity.TargetFile{
		LocalPath:    "dist.zip",
		RemoteName:   "dist.zip",
		UrlPath:      "",
		DefaultCheck: true,
	}
}

func normalizeUrlPath(urlPath string) string {
	trimmed := strings.TrimSpace(urlPath)
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return trimmed
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	escaped := strings.ReplaceAll(value, "'", "'\\''")
	return "'" + escaped + "'"
}

func wrapSudo(cmd string, useSudo bool) string {
	if !useSudo {
		return cmd
	}
	return "sudo sh -c " + shellQuote(cmd)
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
