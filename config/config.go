package config

import (
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
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
	Watcher    Watcher    `yaml:"watcher"`
	Handlers   Handlers   `yaml:"handler"`
}

type Handlers struct {
	CryptoTransferHandler CryptoTransferHandler `yaml:"crypto-transfer"`
}

type CryptoTransferHandler struct {
	TopicId string
}

type Watcher struct {
	CryptoTransfer   CryptoTransfer   `yaml:"crypto-transfer"`
	ConsensusMessage ConsensusMessage `yaml:"consensus-message"`
}

type CryptoTransfer struct {
	Accounts []ID `yaml:"accounts" env:"HEDERA_ETH_BRIDGE_WATCHER_CRYPTO_TRANSFER"`
}

type ConsensusMessage struct {
	Topics []ID `yaml:"topics" env:"HEDERA_ETH_BRIDGE_WATCHER_CONSENSUS_MESSAGE"`
}

type ID struct {
	Id             string `yaml:"id"`
	MaxRetries     int    `yaml:"max_retries"`
	StartTimestamp string `yaml:"start_timestamp"`
}

type Client struct {
	NetworkType string   `yaml:"network_type" env:"HEDERA_ETH_BRIDGE_CLIENT_NETWORK_TYPE"`
	Operator    Operator `yaml:"operator"`
}

type Operator struct {
	AccountId     string `yaml:"account_id" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_ACCOUNT_ID"`
	EthPrivateKey string `yaml:"eth_private_key" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_ETH_PRIVATE_KEY"`
	PrivateKey    string `yaml:"private_key" env:"HEDERA_ETH_BRIDGE_CLIENT_OPERATOR_PRIVATE_KEY"`
}

type MirrorNode struct {
	ClientAddress   string        `yaml:"client_address" env:"HEDERA_ETH_BRIDGE_MIRROR_NODE_CLIENT_ADDRESS"`
	ApiAddress      string        `yaml:"api_address" env:"HEDERA_ETH_BRIDGE_MIRROR_NODE_API_ADDRESS"`
	PollingInterval time.Duration `yaml:"polling_interval" env:"HEDERA_ETH_BRIDGE_MIRROR_NODE_POLLING_INTERVAL"`
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
