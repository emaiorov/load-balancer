package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerConfig struct {
	Url    string `json:"url"`
	Health string `json:"health"`
	Weight uint   `json:"weight"`
}

type Config struct {
	App struct {
		Handler            string `json:"algorythm"`
		Port               string `json:"port"`
		HealthCheckSeconds int    `json:"health_check_seconds"`
	} `json:"app"`
	Servers []ServerConfig `json:"servers"`
}

func Load(path string) (*Config, error) {

	var config Config
	jsonFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file'%s': %w", path, err)
	}

	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON from '%s': %w", path, err)
	}

	return &config, nil
}
