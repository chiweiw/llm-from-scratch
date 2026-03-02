package request

import "deploy-tool/internal/model/entity"

type SaveEnvironment struct {
	Environment entity.Environment `json:"environment"`
}

type DeleteEnvironment struct {
	ID string `json:"id"`
}

type CheckEnvironment struct {
	ID string `json:"id"`
}

type ExportEnvironment struct {
	ID string `json:"id"`
}

type ImportEnvironment struct {
	JSON string `json:"json"`
}

