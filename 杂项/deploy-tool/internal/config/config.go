package config

import _ "embed"

type Config struct {
	ConfigFilePath string
}

//go:embed deploy-tool-config.json
var embeddedDefaultConfig []byte

func Default() Config {
	return Config{ConfigFilePath: "deploy-tool-config.json"}
}

func EmbeddedDefaultConfig() []byte {
	return embeddedDefaultConfig
}

