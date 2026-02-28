package services

import (
	"deploy-tool/internal/models"
)

type DeployService struct {
	progress *models.DeployProgress
}

func NewDeployService() *DeployService {
	return &DeployService{
		progress: &models.DeployProgress{
			Status: "idle",
		},
	}
}

func (s *DeployService) Start(envID string, jarIDs []string) error {
	s.progress = &models.DeployProgress{
		EnvironmentID: envID,
		Status:        "running",
		StartTime:     0,
		TotalProgress: 0,
	}
	return nil
}

func (s *DeployService) Cancel() {
	if s.progress != nil {
		s.progress.Status = "canceled"
	}
}

func (s *DeployService) GetProgress() *models.DeployProgress {
	return s.progress
}
