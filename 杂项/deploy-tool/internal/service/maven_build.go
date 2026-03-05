package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"deploy-tool/internal/logger"
	"deploy-tool/internal/model/entity"
)

type MavenBuildService struct {
	progress      *entity.DeployProgress
	progressMutex sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewMavenBuildService() *MavenBuildService {
	return &MavenBuildService{}
}

type MavenBuildConfig struct {
	ProjectRoot  string
	MavenPath    string
	SettingsPath string
	RepoLocal    string
	Goals        []string
	Properties   map[string]string
	JavaHome     string
	Offline      bool
	Quiet        bool
	UseFPom      bool
}

type MavenBuildResult struct {
	Success      bool
	BuiltFiles   []string
	Duration     time.Duration
	ErrorMessage string
	LogLines     []string
}

func (s *MavenBuildService) StartBuild(ctx context.Context, cfg *MavenBuildConfig, progress *entity.DeployProgress) (*MavenBuildResult, error) {
	logger.Info("开始 Maven 打包，项目根目录: %s", cfg.ProjectRoot)

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.progress = progress
	s.progressMutex.Lock()
	s.progress.Status = entity.DeployStatusRunning
	s.progress.StartTime = time.Now().Unix()
	s.progressMutex.Unlock()

	defer func() {
		s.progressMutex.Lock()
		s.progress.EndTime = time.Now().Unix()
		if s.progress.Status == entity.DeployStatusRunning {
			s.progress.Status = entity.DeployStatusSuccess
		}
		s.progressMutex.Unlock()
	}()

	result := &MavenBuildResult{
		BuiltFiles: []string{},
		LogLines:   []string{},
	}

	cmd, err := s.buildMavenCommand(cfg)
	if err != nil {
		logger.Error("构建 Maven 命令失败: %v", err)
		s.updateStepStatus("Maven 打包", entity.StepStatusFailed, err.Error())
		result.ErrorMessage = err.Error()
		return result, err
	}

	logger.Info("执行 Maven 命令: %s %s", cmd.Path, strings.Join(cmd.Args[1:], " "))
	s.updateStepStatus("Maven 打包", entity.StepStatusRunning, "正在执行 Maven 命令...")

	// 实时流式捕获 stdout / stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return result, fmt.Errorf("获取 stdout pipe 失败: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return result, fmt.Errorf("获取 stderr pipe 失败: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return result, fmt.Errorf("启动 Maven 失败: %v", err)
	}

	// 收集所有日志行（线程安全）
	var logMu sync.Mutex
	var allLines []string

	readLines := func(r io.Reader, isStderr bool) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 支持超长行
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			logMu.Lock()
			allLines = append(allLines, trimmed)
			logMu.Unlock()
			// 根据关键字判断日志级别，实时推送到前端
			lower := strings.ToLower(trimmed)
			switch {
			case strings.Contains(lower, "[error]") || isStderr:
				logger.Error("[Maven] %s", trimmed)
			case strings.Contains(lower, "[warning]") || strings.Contains(lower, "[warn]"):
				logger.Warn("[Maven] %s", trimmed)
			default:
				logger.Info("[Maven] %s", trimmed)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); readLines(stdoutPipe, false) }()
	go func() { defer wg.Done(); readLines(stderrPipe, true) }()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		logger.Error("Maven 构建失败: %v", err)
		s.updateStepStatus("Maven 打包", entity.StepStatusFailed, fmt.Sprintf("Maven 构建失败: %v", err))
		result.ErrorMessage = fmt.Sprintf("Maven 构建失败: %v", err)
		result.LogLines = allLines
		return result, err
	}

	result.LogLines = allLines

	builtFiles, err := s.findBuiltJars(cfg.ProjectRoot, result.LogLines)
	if err != nil {
		logger.Error("查找构建产物失败: %v", err)
		s.updateStepStatus("Maven 打包", entity.StepStatusFailed, err.Error())
		result.ErrorMessage = err.Error()
		return result, err
	}

	logger.Info("构建成功，生成 %d 个 jar 文件", len(builtFiles))
	for _, f := range builtFiles {
		logger.Info("  - %s", f)
	}

	result.BuiltFiles = builtFiles
	result.Duration = time.Since(time.Unix(s.progress.StartTime, 0))

	s.updateStepStatus("Maven 打包", entity.StepStatusSuccess, fmt.Sprintf("构建完成，生成 %d 个文件", len(builtFiles)))
	s.progress.TotalProgress = 100

	result.Success = true
	return result, nil
}

func (s *MavenBuildService) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *MavenBuildService) buildMavenCommand(cfg *MavenBuildConfig) (*exec.Cmd, error) {
	if cfg.ProjectRoot == "" {
		return nil, fmt.Errorf("项目根目录不能为空")
	}

	if !filepath.IsAbs(cfg.ProjectRoot) {
		return nil, fmt.Errorf("项目根目录必须是绝对路径: %s", cfg.ProjectRoot)
	}

	if _, err := os.Stat(cfg.ProjectRoot); os.IsNotExist(err) {
		return nil, fmt.Errorf("项目根目录不存在: %s", cfg.ProjectRoot)
	}

	pomPath := filepath.Join(cfg.ProjectRoot, "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("pom.xml 不存在: %s", pomPath)
	}

	mavenCmd := cfg.MavenPath
	if mavenCmd == "" {
		mavenCmd = s.findMavenExecutable()
		if mavenCmd == "" {
			return nil, fmt.Errorf("未找到 Maven 可执行文件")
		}
	}

	javaHome := strings.TrimSpace(cfg.JavaHome)
	if javaHome == "" {
		javaHome = strings.TrimSpace(os.Getenv("JAVA_HOME"))
	}
	if javaHome == "" {
		return nil, fmt.Errorf("未配置 JDK 路径")
	}
	if !isValidJDKPath(javaHome) {
		return nil, fmt.Errorf("JDK 路径不存在或无效: %s", javaHome)
	}

	args := []string{}

	if cfg.UseFPom {
		args = append(args, "-f", pomPath)
	}

	if cfg.SettingsPath != "" {
		if _, err := os.Stat(cfg.SettingsPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("Maven settings 文件不存在: %s", cfg.SettingsPath)
		}
		args = append(args, "-s", cfg.SettingsPath)
	}

	if cfg.RepoLocal != "" {
		args = append(args, fmt.Sprintf("-Dmaven.repo.local=%s", cfg.RepoLocal))
	}

	if cfg.Offline {
		args = append(args, "-o")
	}

	if cfg.Quiet {
		args = append(args, "-q")
	}

	for key, value := range cfg.Properties {
		args = append(args, fmt.Sprintf("-D%s=%s", key, value))
	}

	if len(cfg.Goals) == 0 {
		args = append(args, "clean", "package", "-DskipTests")
	} else {
		args = append(args, cfg.Goals...)
	}

	cmd := exec.CommandContext(s.ctx, mavenCmd, args...)
	cmd.Dir = cfg.ProjectRoot

	// Windows 下隐藏控制台窗口，防止弹出 cmd 窗口被用户误关导致进程终止
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
			HideWindow:    true,
		}
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("JAVA_HOME=%s", javaHome))
	javaBin := filepath.Join(javaHome, "bin")
	env = prependPath(env, javaBin)
	cmd.Env = env

	return cmd, nil
}

func (s *MavenBuildService) findMavenExecutable() string {
	var candidates []string

	if runtime.GOOS == "windows" {
		candidates = []string{"mvn.cmd", "mvn.bat"}
	} else {
		candidates = []string{"mvn"}
	}

	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	return ""
}

func isValidJDKPath(javaHome string) bool {
	if runtime.GOOS == "windows" {
		javaCmd := filepath.Join(javaHome, "bin", "java.exe")
		javacCmd := filepath.Join(javaHome, "bin", "javac.exe")
		return fileExists(javaCmd) || fileExists(javacCmd)
	}
	javaCmd := filepath.Join(javaHome, "bin", "java")
	javacCmd := filepath.Join(javaHome, "bin", "javac")
	return fileExists(javaCmd) || fileExists(javacCmd)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func prependPath(env []string, value string) []string {
	if value == "" {
		return env
	}
	sep := string(os.PathListSeparator)
	for i, kv := range env {
		if strings.HasPrefix(strings.ToUpper(kv), "PATH=") {
			env[i] = "PATH=" + value + sep + strings.TrimPrefix(kv, "PATH=")
			return env
		}
	}
	return append(env, "PATH="+value)
}

func (s *MavenBuildService) findBuiltJars(projectRoot string, logLines []string) ([]string, error) {
	var jars []string

	jarPattern := regexp.MustCompile(`Building jar: (.+\.jar)`)
	for _, line := range logLines {
		if matches := jarPattern.FindStringSubmatch(line); len(matches) > 1 {
			jarPath := strings.TrimSpace(matches[1])
			if filepath.IsAbs(jarPath) {
				jars = append(jars, jarPath)
			} else {
				absPath := filepath.Join(projectRoot, jarPath)
				jars = append(jars, absPath)
			}
		}
	}

	if len(jars) == 0 {
		err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".jar") {
				if strings.Contains(path, "target") && !strings.Contains(path, "original-") {
					jars = append(jars, path)
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("扫描 jar 文件失败: %v", err)
		}
	}

	return jars, nil
}

func (s *MavenBuildService) parseLogLines(content string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func (s *MavenBuildService) updateStepStatus(stepName string, status string, message string) {
	s.progressMutex.Lock()
	defer s.progressMutex.Unlock()

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

func (s *MavenBuildService) updateProgress(percent int) {
	s.progressMutex.Lock()
	defer s.progressMutex.Unlock()

	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}
	s.progress.TotalProgress = percent
}
