package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	Port    string `yaml:"port,omitempty"`
	BaseURL string `yaml:"basePath,omitempty"`
}

type Config struct {
	Service ServiceConfig `yaml:"service,omitempty"`
}

func NewConfig(configPath *string) (*Config, error) {
	file, err := os.Open(*configPath)
	if err != nil {
		err = fmt.Errorf("failed to open config file: %w", err)
		return nil, err
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("failed to read bytes from config file: %w", err)
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		err = fmt.Errorf("failed to unmarshall config file: %w", err)
		return nil, err
	}

	if config.Service.Port == "" {
		config.Service.Port = "8080"
	}

	if config.Service.BaseURL == "" {
		config.Service.BaseURL = "api/v1"
	}

	return &config, nil
}
