package config

import (
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultConfigFile = "config/application.yml"
)

func LoadConfig() *Config {
	var configuration Config
	GetConfig(&configuration, defaultConfigFile)

	if err := env.Parse(&configuration); err != nil {
		panic(err)
	}

	return &configuration
}

func GetConfig(config *Config, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}

	filename, _ := filepath.Abs(path)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	Hedera Hedera `yaml:"hedera"`
}

type Hedera struct {
	Validator Validator `yaml:"validator"`
}

type Validator struct {
	Db   Db     `yaml:"db"`
	Port string `yaml:"port" env:"HEDERA_ETH_BRIDGE_VALIDATOR_PORT"`
}

type Db struct {
	Host     string `yaml:"host" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_HOST"`
	Name     string `yaml:"name" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_NAME"`
	Password string `yaml:"password" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_PASSWORD"`
	Port     string `yaml:"port" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_PORT"`
	Username string `yaml:"username" env:"HEDERA_ETH_BRIDGE_VALIDATOR_DB_USERNAME"`
}
