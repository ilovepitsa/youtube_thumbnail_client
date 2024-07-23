package config

import (
	"path"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Network `yaml:"network"`
		Logger  `yaml:"log"`
	}

	Network struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	}
	Logger struct {
		Level  string `yaml:"level"`
		IsJSON bool   `yaml:"is_json"`
	}
)

func NewConfig(pathToConfig string) (*Config, error) {
	config := &Config{}
	err := cleanenv.ReadConfig(path.Join("./", pathToConfig), config)
	if err != nil {
		return nil, err
	}

	err = cleanenv.UpdateEnv(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
