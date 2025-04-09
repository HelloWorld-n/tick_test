package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	BaseURL string `yaml:"baseURL"`
	Port    string `yaml:"port"`
}

var config *Config = nil

func GetConfig(path string) (cfg *Config, err error) {
	if config != nil {
		return
	}
	cfg = &Config{
		Port: "4041",
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, cfg)
	return
}
