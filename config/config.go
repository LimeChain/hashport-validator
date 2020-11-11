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
	Validator  Validator  `yaml:"validator"`
	MirrorNode MirrorNode `yaml:"mirror_node"`
	Client     Client     `yaml:"client"`
}

type Client struct {
	NetworkType string   `yaml:"network_type" env:"HEDERA_ETH_BRIDGE_CLIENT_NETWORK_TYPE"`
	Operator    Operator `yaml:"operator"`
}

type Operator struct {
	AccountId  string `yaml:"account_id" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_ACCOUNT_ID"`
	PublicKey  string `yaml:"public_key" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_PUBLIC_KEY"`
	PrivateKey string `yaml:"private_key" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_PRIVATE_KEY"`
}

type MirrorNode struct {
	Client     string `yaml:"client" env:"HEDERA_ETH_BRIDGE_MIRROR_NODE_CLIENT"`
	ApiAddress string `yaml:"api_address" env:"HEDERA_ETH_BRIDGE_MIRROR_NODE_API_ADDRESS"`
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
