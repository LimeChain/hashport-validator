package config

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_New(t *testing.T) {
	in := parser.Node{
		Database: parser.Database{
			Host:     "db-host",
			Name:     "db-name",
			Password: "db-pass",
			Port:     "db-port",
			Username: "db-user",
		},
		Clients: parser.Clients{
			Evm: map[uint64]parser.Evm{
				80001: {
					BlockConfirmations: 1,
					NodeUrl:            "node-url",
					PrivateKey:         "private-key",
					StartBlock:         0,
					PollingInterval:    0,
					MaxLogsBlocks:      0,
				},
			},
			Hedera: parser.Hedera{
				Operator: parser.Operator{
					AccountId:  "account-id",
					PrivateKey: "private-key",
				},
				Network:        "network",
				StartTimestamp: 0,
			},
			MirrorNode: parser.MirrorNode{
				ClientAddress:   "client-address",
				ApiAddress:      "api-address",
				PollingInterval: 0,
			},
		},
		LogLevel:  "log-level",
		Port:      "port",
		Validator: false,
		Monitoring: parser.Monitoring{
			Enable:           false,
			DashboardPolling: 0,
		},
	}

	expected := Node{
		Database: Database{
			Host:     "db-host",
			Name:     "db-name",
			Password: "db-pass",
			Port:     "db-port",
			Username: "db-user",
		},
		Clients: Clients{
			Evm: map[uint64]Evm{
				80001: {
					BlockConfirmations: 1,
					NodeUrl:            "node-url",
					PrivateKey:         "private-key",
					StartBlock:         0,
					PollingInterval:    0,
					MaxLogsBlocks:      0,
				},
			},
			Hedera: Hedera{
				Operator: Operator{
					AccountId:  "account-id",
					PrivateKey: "private-key",
				},
				Network:        "network",
				StartTimestamp: 0,
				Rpc:            map[string]hedera.AccountID{},
			},
			MirrorNode: MirrorNode{
				ClientAddress:   "client-address",
				ApiAddress:      "api-address",
				PollingInterval: 0,
			},
		},
		LogLevel:  "log-level",
		Port:      "port",
		Validator: false,
		Monitoring: Monitoring{
			Enable:           false,
			DashboardPolling: 0,
		},
	}

	actual := New(in)
	assert.Equal(t, actual, expected)
}

func Test_parseRpc(t *testing.T) {
	acc1, _ := hedera.AccountIDFromString("0.0.1")
	acc2, _ := hedera.AccountIDFromString("0.0.2")
	acc3, _ := hedera.AccountIDFromString("0.0.3")

	expected := map[string]hedera.AccountID{
		"key1": acc1,
		"key2": acc2,
		"key3": acc3,
	}

	in := map[string]string{
		"key1": "0.0.1",
		"key2": "0.0.2",
		"key3": "0.0.3",
	}

	actual := parseRpc(in)
	assert.Equal(t, expected, actual)
}
