package request

import "deploy-tool/internal/model/entity"

type SaveGlobalSettings struct {
	Settings entity.GlobalSettings `json:"settings"`
}

type SaveSystemDefaults struct {
	Defaults entity.SystemDefaultConfig `json:"defaults"`
}

type SetLastSelectedEnvID struct {
	ID string `json:"id"`
}

