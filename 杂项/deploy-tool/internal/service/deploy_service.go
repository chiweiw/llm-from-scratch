package service

import "deploy-tool/internal/model/entity"

type DeployService struct {
	progress *entity.DeployProgress
}

func NewDeployService() *DeployService {
	return &DeployService{
		progress: &entity.DeployProgress{
			Status: entity.DeployStatusIdle,
		},
	}
}

func (s *DeployService) Start(envID string, jarIDs []string) error {
	s.progress = &entity.DeployProgress{
		EnvironmentID: envID,
		Status:        entity.DeployStatusRunning,
		StartTime:     0,
		TotalProgress: 0,
	}
	return nil
}

func (s *DeployService) Cancel() {
	if s.progress != nil {
		s.progress.Status = entity.DeployStatusCanceled
	}
}

func (s *DeployService) GetProgress() *entity.DeployProgress {
	return s.progress
}

