package main

import (
	yaml "gopkg.in/yaml.v3"
	"os"
	"pi.com/lb/app"
	"pi.com/lb/model"
)

const (
	DEFAULT_CONFIG_FILE_NAME = "config.yaml"
)

func main() {
	fileName := DEFAULT_CONFIG_FILE_NAME

	cfg, err := readConfigFile(fileName)
	if err != nil {
		panic(err)
	}

	err = app.NewLBApp(cfg)
	if err != nil {
		panic(err)
	}
}

func readConfigFile(fileName string) (*model.Config, error) {
	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	cfg := &model.Config{}
	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil

}
