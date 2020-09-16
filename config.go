package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gopkg.in/yaml.v3"
)

// Config struct
type Config struct {
	Logger struct {
		Level        string        `yaml:"level"`
		LevelEncoded zapcore.Level `yaml:"levelEncoded,omitempty"`
	} `yaml:"logger"`
	VictoriaMetrics string `yaml:"victoriaMetrics"`
	AppPort         string `yaml:"appPort"`
}

var config Config
var logger *zap.Logger

func makeConfig() {
	configPath := flag.String("config", "", "path2ConfigFile")
	flag.Parse()
	if *configPath == "" {
		log.Fatalln("VictoriaMetrics host must be specified!\n Example: ./myApp --config /path/to/config.yaml")
	}
	decodeConfig(configPath)
	parseConfig()

}

func parseConfig() {
	if config.Logger.Level == "" || config.Logger.Level == "info" {
		config.Logger.LevelEncoded = 0
	} else if config.Logger.Level == "debug" {
		config.Logger.LevelEncoded = -1
	} else {
		config.Logger.LevelEncoded = 1
	}

	if config.VictoriaMetrics == "" {
		log.Fatalln("VictoriaMetrics host must be specified!\n Example: VictoriaMetrics: http://VictoriaMetrics:8428")
	}

	if config.AppPort == "" {
		config.AppPort = "8080"
		config.AppPort = fmt.Sprintf(":%v", config.AppPort)
	} else {
		config.AppPort = fmt.Sprintf(":%v", config.AppPort)
	}

}

func decodeConfig(configPath *string) {
	f, err := os.Open(*configPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	configEncoder := yaml.NewDecoder(f)
	err = configEncoder.Decode(&config)
	if err != nil {
		log.Fatalln(err)
	}

}
