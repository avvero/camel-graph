package model

import (
	"os"
	"encoding/json"
	"time"
)

type InstanceConfig struct {
	Created                      time.Time
	Environments                 []*EnvironmentConfig
	ServiceUpdateIntervalSeconds int
	RouteUpdateIntervalSeconds   int
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

type Authorization struct {
	Login string
	Pass  string
}

// Read config
func ReadConfig(fileName string) (*InstanceConfig, error) {
	rootConfig := InstanceConfig{}
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
