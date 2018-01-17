package main

import (
	"os"
	"encoding/json"
	"time"
)

type Authorization struct {
	Login string
	Pass  string
}

type Config struct {
	Created      time.Time
	Environments []*EnvironmentConfig
}

type EnvironmentConfig struct {
	Name     string
	Services []*ServiceConfig
}

type ServiceConfig struct {
	Name          string
	Url           string
	Color         string
	Authorization *Authorization
}

// Read config
func ReadConfig(fileName string) (*Config, error) {
	rootConfig := Config{}
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&rootConfig)
	if err != nil {
		return nil, err
	}
	return &rootConfig, nil
}
