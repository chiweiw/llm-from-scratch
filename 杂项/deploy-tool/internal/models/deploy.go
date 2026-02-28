package models

type DeployProgress struct {
	EnvironmentID  string       `json:"environmentId"`
	Status         string       `json:"status"`
	CurrentStep    string       `json:"currentStep"`
	TotalProgress  int          `json:"totalProgress"`
	Steps          []StepProgress `json:"steps"`
	CurrentFile    string       `json:"currentFile"`
	FileProgress   int          `json:"fileProgress"`
	Speed          string       `json:"speed"`
	StartTime      int64        `json:"startTime"`
	EndTime        int64        `json:"endTime"`
	ErrorMessage   string       `json:"errorMessage"`
}

type StepProgress struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
}

const (
	DeployStatusIdle     = "idle"
	DeployStatusRunning  = "running"
	DeployStatusSuccess  = "success"
	DeployStatusFailed   = "failed"
	DeployStatusCanceled = "canceled"
)

const (
	StepStatusPending = "pending"
	StepStatusRunning = "running"
	StepStatusSuccess = "success"
	StepStatusFailed  = "failed"
	StepStatusSkipped = "skipped"
)
