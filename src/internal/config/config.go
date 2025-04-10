package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	BaseURL string `yaml:"baseURL"`
	Port    string `yaml:"port"`
}

func GetConfig(path string) (cfg *Config, err error) {
	data, err := os.ReadFile(path)
	cfg = &Config{
		Port: "4041",
	}
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, cfg)
	return
}
