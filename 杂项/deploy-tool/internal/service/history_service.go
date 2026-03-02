package service

import "deploy-tool/internal/model/entity"

type HistoryService struct {
	records []entity.DeployHistory
}

func NewHistoryService() *HistoryService {
	return &HistoryService{
		records: make([]entity.DeployHistory, 0),
	}
}

func (s *HistoryService) GetList() []entity.DeployHistory {
	return s.records
}

func (s *HistoryService) GetDetail(id string) *entity.DeployHistory {
	for i := range s.records {
		if s.records[i].ID == id {
			return &s.records[i]
		}
	}
	return nil
}

func (s *HistoryService) Add(record entity.DeployHistory) {
	s.records = append(s.records, record)
}

