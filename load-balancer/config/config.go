package config

import (
	"encoding/json"
	"log"
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

func Load() *Config {

	var config Config
	jsonFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	return &config
}
