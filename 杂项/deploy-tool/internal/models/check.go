package models

type CheckResult struct {
	Success bool        `json:"success"`
	Checks  []CheckItem `json:"checks"`
	Summary string      `json:"summary"`
}

type CheckItem struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

const (
	CheckStatusPass    = "pass"
	CheckStatusFail    = "error"
	CheckStatusWarning = "warning"
)
