package entity

type DeployHistory struct {
	ID              string   `json:"id"`
	EnvironmentID   string   `json:"environmentId"`
	EnvironmentName string   `json:"environmentName"`
	StartTime       int64    `json:"startTime"`
	EndTime         int64    `json:"endTime"`
	Status          string   `json:"status"`
	Files           []string `json:"files"`
	Duration        int64    `json:"duration"`
	ErrorMessage    string   `json:"errorMessage"`
}

type HistoryFilter struct {
	EnvironmentID string `json:"environmentId"`
	Status        string `json:"status"`
	StartDate     int64  `json:"startDate"`
	EndDate       int64  `json:"endDate"`
	Page          int    `json:"page"`
	PageSize      int    `json:"pageSize"`
}

type DeployLog struct {
	ID        string `json:"id"`
	DeployID  string `json:"deployId"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	CreatedAt int64  `json:"createdAt"`
}

