package main

import (
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
	"os"
	"pi.com/lb/app"
	"pi.com/lb/model"
)

const (
	DEFAULT_CONFIG_FILE_NAME = "config/config.yaml"
)

func main() {
	fileName := DEFAULT_CONFIG_FILE_NAME

	logger := logrus.New()
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logger.SetFormatter(customFormatter)

	logger.SetLevel(logrus.DebugLevel)

	cfg, err := readConfigFile(fileName)
	if err != nil {
		logger.Fatal(err)
	}

	err = app.NewLBApp(cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}
}

func readConfigFile(fileName string) (*model.LoadBalancerConfig, error) {
	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	cfg := &model.LoadBalancerConfig{}
	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil

}
