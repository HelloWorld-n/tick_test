package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

const configFile = "../config.yaml"

type Config struct {
	BaseURL string `yaml:"baseURL"`
	Port    string `yaml:"port"`
}

var config *Config = nil

func GetConfig() (cfg *Config, err error) {
	if config != nil {
		return
	}

	file, err := os.Open(configFile)
	if err != nil {
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	cfg = new(Config)
	err = yaml.Unmarshal(data, &cfg)
	return
}
