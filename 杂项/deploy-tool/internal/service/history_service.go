package service

import (
	"deploy-tool/internal/db"
	"deploy-tool/internal/model/entity"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type HistoryService struct {
	historyDAO db.DeployHistoryDAO
	logDAO     db.DeployLogDAO
}

func NewHistoryService(historyDAO db.DeployHistoryDAO, logDAO db.DeployLogDAO) *HistoryService {
	return &HistoryService{
		historyDAO: historyDAO,
		logDAO:     logDAO,
	}
}

func (s *HistoryService) GetAll(limit int) ([]entity.DeployHistory, error) {
	dbHistories, err := s.historyDAO.GetAll(limit)
	if err != nil {
		return nil, err
	}

	histories := make([]entity.DeployHistory, 0, len(dbHistories))
	for _, dbHistory := range dbHistories {
		files := parseFilesField(dbHistory.Files)
		histories = append(histories, entity.DeployHistory{
			ID:              dbHistory.ID,
			EnvironmentID:   dbHistory.EnvironmentID,
			EnvironmentName: dbHistory.EnvironmentName,
			StartTime:       dbHistory.StartTime,
			EndTime:         dbHistory.EndTime,
			Status:          dbHistory.Status,
			Files:           files,
			Duration:        dbHistory.Duration,
			ErrorMessage:    dbHistory.ErrorMessage,
		})
	}
	return histories, nil
}

func (s *HistoryService) GetByEnvironmentID(envID string, limit int) ([]entity.DeployHistory, error) {
	dbHistories, err := s.historyDAO.GetByEnvironmentID(envID, limit)
	if err != nil {
		return nil, err
	}

	histories := make([]entity.DeployHistory, 0, len(dbHistories))
	for _, dbHistory := range dbHistories {
		files := parseFilesField(dbHistory.Files)
		histories = append(histories, entity.DeployHistory{
			ID:              dbHistory.ID,
			EnvironmentID:   dbHistory.EnvironmentID,
			EnvironmentName: dbHistory.EnvironmentName,
			StartTime:       dbHistory.StartTime,
			EndTime:         dbHistory.EndTime,
			Status:          dbHistory.Status,
			Files:           files,
			Duration:        dbHistory.Duration,
			ErrorMessage:    dbHistory.ErrorMessage,
		})
	}
	return histories, nil
}

func (s *HistoryService) GetByID(id string) (*entity.DeployHistory, error) {
	dbHistory, err := s.historyDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	files := parseFilesField(dbHistory.Files)
	result := &entity.DeployHistory{
		ID:              dbHistory.ID,
		EnvironmentID:   dbHistory.EnvironmentID,
		EnvironmentName: dbHistory.EnvironmentName,
		StartTime:       dbHistory.StartTime,
		EndTime:         dbHistory.EndTime,
		Status:          dbHistory.Status,
		Files:           files,
		Duration:        dbHistory.Duration,
		ErrorMessage:    dbHistory.ErrorMessage,
	}
	return result, nil
}

func (s *HistoryService) Create(envID, envName string) string {
	now := time.Now().Unix()
	historyID := fmt.Sprintf("deploy_%d", now)

	dbHistory := &db.DeployHistory{
		ID:              historyID,
		EnvironmentID:   envID,
		EnvironmentName: envName,
		StartTime:       now,
		Status:          "running",
	}

	if err := s.historyDAO.Create(dbHistory); err != nil {
		return ""
	}

	return historyID
}

func (s *HistoryService) Update(historyID string, status string, files string, duration int64, errorMsg string) error {
	now := time.Now().Unix()
	dbHistory := &db.DeployHistory{
		ID:           historyID,
		EndTime:      now,
		Status:       status,
		Files:        files,
		Duration:     duration,
		ErrorMessage: errorMsg,
	}
	return s.historyDAO.Update(dbHistory)
}

func (s *HistoryService) Delete(id string) error {
	return s.historyDAO.Delete(id)
}

func (s *HistoryService) DeleteOld(days int) error {
	return s.historyDAO.DeleteOld(days)
}

func (s *HistoryService) AddLog(historyID, level, message string) error {
	now := time.Now().Unix()
	log := &db.DeployLog{
		ID:        fmt.Sprintf("log_%d_%d", now, time.Now().UnixNano()),
		DeployID:  historyID,
		Level:     level,
		Message:   message,
		Timestamp: now,
		CreatedAt: now,
	}
	return s.logDAO.Create(log)
}

func (s *HistoryService) GetLogs(historyID string) ([]db.DeployLog, error) {
	return s.logDAO.GetByDeployID(historyID)
}

func parseFilesField(filesText string) []string {
	if filesText == "" {
		return []string{}
	}
	// Try JSON array first
	var arr []string
	if err := json.Unmarshal([]byte(filesText), &arr); err == nil {
		return arr
	}
	// Fallback: comma-separated
	parts := strings.Split(filesText, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
