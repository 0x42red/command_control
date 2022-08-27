package embeddata

import (
	"embed"

	yaml "gopkg.in/yaml.v3"
)

//go:embed config/config.yaml
var configFile embed.FS

type Config struct {
	Port int
	Host string
}

func GetConfig() (*Config, error) {
	b, err := configFile.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}

	data := &Config{}

	err = yaml.Unmarshal(b, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
