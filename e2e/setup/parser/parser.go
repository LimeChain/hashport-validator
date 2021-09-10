package parser

import (
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
)

// Config used to load and parse from application.yml
type Config struct {
	Hedera       HederaParser         `yaml:"hedera"`
	EVM          map[int64]parser.Evm `yaml:"evm"`
	Tokens       Tokens               `yaml:"tokens"`
	ValidatorUrl string               `yaml:"validator_url"`
	Bridge       parser.Bridge        `yaml:"bridge"`
}

type HederaParser struct {
	NetworkType       string            `yaml:"network_type"`
	BridgeAccount     string            `yaml:"bridge_account"`
	Members           []string          `yaml:"members"`
	TopicID           string            `yaml:"topic_id"`
	Sender            Sender            `yaml:"sender"`
	DbValidationProps []parser.Database `yaml:"dbs"`
	MirrorNode        parser.MirrorNode `yaml:"mirror_node"`
}

type Sender struct {
	Account    string `yaml:"account"`
	PrivateKey string `yaml:"private_key"`
}

type Tokens struct {
	WHbar          string `yaml:"whbar"`
	WToken         string `yaml:"wtoken"`
	EvmNativeToken string `yaml:"evm_native_token"`
}

type E2E struct {
	Hedera       HederaParser         `yaml:"hedera"`
	EVM          map[int64]parser.Evm `yaml:"evm"`
	Tokens       Tokens               `yaml:"tokens"`
	ValidatorUrl string               `yaml:"validator_url"`
	Bridge       parser.Bridge
}
