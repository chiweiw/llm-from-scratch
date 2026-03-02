package config

type Config struct {
	ConfigFilePath string
}

func Default() Config {
	return Config{ConfigFilePath: "deploy-tool-config.json"}
}

