package main

import (
	"encoding/json"
	"fmt"

	"mongodriver"
	"os"
	"path"
	"runtime"
)

// Configuration is struct collect all configuration
type Configuration struct {
	Mongo mongodriver.Mongo `json:"mongo_srv"`
}

// CentralizeConfiguration object
type CentralizeConfiguration struct {
	Mongo mongodriver.Mongo `json:"mongo_srv"`
}

// GetConfiguration load config from file
func GetConfiguration() Configuration {
	_, filename, _, _ := runtime.Caller(1)
	configuration := Configuration{}
	configFile, _ := os.Open(path.Join(path.Dir(filename), "config/db.json"))
	DecodeConfig(configFile, &configuration)
	return configuration
}

// DecodeConfig is decode configuration
func DecodeConfig(file *os.File, target interface{}) {
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&target); err != nil {
		panic(fmt.Sprintf("Cannot decode file: %s", err))
	}
}
