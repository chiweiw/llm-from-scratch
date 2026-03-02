package request

type StartDeploy struct {
	EnvironmentID string   `json:"environmentId"`
	JarIDs        []string `json:"jarIds"`
}

