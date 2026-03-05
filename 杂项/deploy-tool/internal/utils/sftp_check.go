package utils

import (
	"deploy-tool/internal/model/entity"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPClient struct {
	client *ssh.Client
}

func NewSFTPClient(host string, port int, username, password string) (*SFTPClient, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %v", err)
	}

	return &SFTPClient{client: client}, nil
}

func (s *SFTPClient) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *SFTPClient) UploadFile(localPath, remoteDir, remoteName string) error {
	if s.client == nil {
		return fmt.Errorf("SSH客户端未初始化")
	}

	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	if localPath == "" {
		return fmt.Errorf("本地文件路径为空")
	}

	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("本地文件不存在或无法访问: %v", err)
	}
	if info.IsDir() {
		return fmt.Errorf("暂不支持上传目录: %s", localPath)
	}

	if remoteDir == "" {
		return fmt.Errorf("远程部署目录未配置")
	}

	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("创建远程目录失败: %v", err)
	}

	if remoteName == "" {
		remoteName = filepath.Base(localPath)
	}

	remotePath := path.Join(remoteDir, remoteName)

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer src.Close()

	dst, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}

	return nil
}

func (s *SFTPClient) UploadFileWithProgress(localPath, remoteDir, remoteName string, onProgress func(written, total int64)) error {
	if s.client == nil {
		return fmt.Errorf("SSH客户端未初始化")
	}

	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	if localPath == "" {
		return fmt.Errorf("本地文件路径为空")
	}

	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("本地文件不存在或无法访问: %v", err)
	}
	if info.IsDir() {
		return fmt.Errorf("暂不支持上传目录: %s", localPath)
	}
	total := info.Size()

	if remoteDir == "" {
		return fmt.Errorf("远程部署目录未配置")
	}
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("创建远程目录失败: %v", err)
	}

	if remoteName == "" {
		remoteName = filepath.Base(localPath)
	}
	remotePath := path.Join(remoteDir, remoteName)

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer src.Close()

	dst, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %v", err)
	}
	defer dst.Close()

	const bufSize = 256 * 1024
	buf := make([]byte, bufSize)
	var written int64 = 0
	if onProgress != nil {
		onProgress(0, total)
	}
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			w, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("写入远程文件失败: %v", writeErr)
			}
			if w != n {
				return fmt.Errorf("写入远程文件字节数不一致: %d != %d", w, n)
			}
			written += int64(w)
			if onProgress != nil {
				onProgress(written, total)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("读取本地文件失败: %v", readErr)
		}
	}
	return nil
}

func (s *SFTPClient) CheckConnection() error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	err = session.Run("echo connection ok")
	if err != nil {
		return fmt.Errorf("SSH连接测试失败: %v", err)
	}

	return nil
}

func (s *SFTPClient) CheckDeployDir(deployDir string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("test -d %s && echo exists || echo notexists", deployDir)
	output, err := session.Output(cmd)
	if err != nil {
		return fmt.Errorf("检查部署目录失败: %v", err)
	}

	if string(output) == "notexists\n" {
		return fmt.Errorf("部署目录不存在: %s", deployDir)
	}

	return nil
}

func (s *SFTPClient) TestRestartScript(scriptPath string) error {
	if scriptPath == "" {
		return nil
	}

	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("test -f %s && echo exists || echo notexists", scriptPath)
	output, err := session.Output(cmd)
	if err != nil {
		return fmt.Errorf("检查重启脚本失败: %v", err)
	}

	if string(output) == "notexists\n" {
		return fmt.Errorf("重启脚本不存在: %s", scriptPath)
	}

	return nil
}

func (s *SFTPClient) RunCommand(cmd string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("SSH客户端未初始化")
	}

	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("执行远程命令失败: %v", err)
	}

	return string(output), nil
}

func (s *SFTPClient) BackupRemoteFile(remoteDir, remoteName string, cleanup bool) error {
	if s.client == nil {
		return fmt.Errorf("SSH客户端未初始化")
	}
	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("创建SFTP客户端失败: %v", err)
	}
	defer sftpClient.Close()

	if remoteDir == "" || remoteName == "" {
		return fmt.Errorf("远程目录或文件名为空")
	}
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("创建远程目录失败: %v", err)
	}

	base := path.Join(remoteDir, remoteName)
	_, statErr := sftpClient.Stat(base)
	if statErr != nil {
		// 文件不存在，无需备份
		return nil
	}

	if cleanup {
		dirList, err := sftpClient.ReadDir(remoteDir)
		if err == nil {
			prefix := remoteName + "."
			for _, f := range dirList {
				name := f.Name()
				if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".bak") {
					_ = sftpClient.Remove(path.Join(remoteDir, name))
				}
			}
		}
	}

	ts := time.Now().Format("20060102150405")
	bak := base + "." + ts + ".bak"
	if err := sftpClient.Rename(base, bak); err != nil {
		return fmt.Errorf("远程备份失败: %v", err)
	}
	return nil
}

func CheckSFTPServer(server *entity.ServerConfig) *entity.CheckItem {
	if server.Host == "" || server.Port == 0 || server.Username == "" {
		return &entity.CheckItem{
			Name:    "SFTP服务器: " + server.Name,
			Status:  entity.CheckStatusFail,
			Message: "服务器配置不完整",
		}
	}

	client, err := NewSFTPClient(server.Host, server.Port, server.Username, server.Password)
	if err != nil {
		return &entity.CheckItem{
			Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
			Status:  entity.CheckStatusFail,
			Message: fmt.Sprintf("无法连接: %v", err),
		}
	}
	defer client.Close()

	if err := client.CheckConnection(); err != nil {
		return &entity.CheckItem{
			Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
			Status:  entity.CheckStatusFail,
			Message: fmt.Sprintf("连接失败: %v", err),
		}
	}

	if server.DeployDir != "" {
		if err := client.CheckDeployDir(server.DeployDir); err != nil {
			return &entity.CheckItem{
				Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
				Status:  entity.CheckStatusWarning,
				Message: fmt.Sprintf("警告: %v", err),
			}
		}
	}

	if server.RestartScript != "" {
		if err := client.TestRestartScript(server.RestartScript); err != nil {
			return &entity.CheckItem{
				Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
				Status:  entity.CheckStatusWarning,
				Message: fmt.Sprintf("重启脚本: %v", err),
			}
		}
	}

	return &entity.CheckItem{
		Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
		Status:  entity.CheckStatusPass,
		Message: fmt.Sprintf("%s@%s:%d", server.Username, server.Host, server.Port),
	}
}
