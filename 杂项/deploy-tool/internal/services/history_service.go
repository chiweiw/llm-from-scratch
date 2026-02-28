package services

import (
	"deploy-tool/internal/models"
)

type HistoryService struct {
	records []models.DeployHistory
}

func NewHistoryService() *HistoryService {
	return &HistoryService{
		records: make([]models.DeployHistory, 0),
	}
}

func (s *HistoryService) GetList() []models.DeployHistory {
	return s.records
}

func (s *HistoryService) GetDetail(id string) *models.DeployHistory {
	for i := range s.records {
		if s.records[i].ID == id {
			return &s.records[i]
		}
	}
	return nil
}

func (s *HistoryService) Add(record models.DeployHistory) {
	s.records = append(s.records, record)
}
