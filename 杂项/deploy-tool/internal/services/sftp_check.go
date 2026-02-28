package services

import (
	"deploy-tool/internal/models"
	"fmt"
	"time"

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

func CheckSFTPServer(server *models.ServerConfig) *models.CheckItem {
	if server.Host == "" || server.Port == 0 || server.Username == "" {
		return &models.CheckItem{
			Name:    "SFTP服务器: " + server.Name,
			Status:  models.CheckStatusFail,
			Message: "服务器配置不完整",
		}
	}

	client, err := NewSFTPClient(server.Host, server.Port, server.Username, server.Password)
	if err != nil {
		return &models.CheckItem{
			Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
			Status:  models.CheckStatusFail,
			Message: fmt.Sprintf("无法连接: %v", err),
		}
	}
	defer client.Close()

	if err := client.CheckConnection(); err != nil {
		return &models.CheckItem{
			Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
			Status:  models.CheckStatusFail,
			Message: fmt.Sprintf("连接失败: %v", err),
		}
	}

	if server.DeployDir != "" {
		if err := client.CheckDeployDir(server.DeployDir); err != nil {
			return &models.CheckItem{
				Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
				Status:  models.CheckStatusWarning,
				Message: fmt.Sprintf("警告: %v", err),
			}
		}
	}

	if server.RestartScript != "" {
		if err := client.TestRestartScript(server.RestartScript); err != nil {
			return &models.CheckItem{
				Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
				Status:  models.CheckStatusWarning,
				Message: fmt.Sprintf("重启脚本: %v", err),
			}
		}
	}

	return &models.CheckItem{
		Name:    "SFTP服务器: " + server.Name + " (" + server.Host + ")",
		Status:  models.CheckStatusPass,
		Message: fmt.Sprintf("%s@%s:%d", server.Username, server.Host, server.Port),
	}
}
