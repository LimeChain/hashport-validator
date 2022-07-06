/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	log "github.com/sirupsen/logrus"
)

type Node struct {
	Database   Database
	Clients    Clients
	LogLevel   string
	Port       string
	Validator  bool
	Monitoring Monitoring
}

type Database struct {
	Host     string
	Name     string
	Password string
	Port     string
	Username string
}

type Clients struct {
	Evm           map[uint64]Evm
	Hedera        Hedera
	MirrorNode    MirrorNode
	CoinGecko     CoinGecko
	CoinMarketCap CoinMarketCap
}

type Evm struct {
	BlockConfirmations uint64
	NodeUrl            string
	PrivateKey         string
	StartBlock         int64
	PollingInterval    time.Duration
	MaxLogsBlocks      int64
}

type Hedera struct {
	Operator       Operator
	Network        string
	Rpc            map[string]hedera.AccountID
	StartTimestamp int64
}

type Operator struct {
	AccountId  string
	PrivateKey string
}

// CoinGecko //

type CoinGecko struct {
	ApiAddress string
}

// CoinMarketCap //

type CoinMarketCap struct {
	ApiKey     string
	ApiAddress string
}

// MirrorNode //

type MirrorNode struct {
	ClientAddress     string
	ApiAddress        string
	PollingInterval   time.Duration
	QueryMaxLimit     int64
	QueryDefaultLimit int64
	RetryPolicy       RetryPolicy
	RequestTimeout    time.Duration
}

const (
	// in seconds
	defaultPollingInterval   = 5
	defaultQueryMaxLimit     = 100
	defaultQueryDefaultLimit = 25
	// in seconds
	defaultRequestTimeout = 15
)

func (m *MirrorNode) DefaultOrConfig(cfg *parser.MirrorNode) *MirrorNode {
	if cfg.ClientAddress == "" {
		log.Fatalf("node configuration: MirrorNode ClientAddress is required")
	}
	m.ClientAddress = cfg.ClientAddress
	if cfg.ApiAddress == "" {
		log.Fatalf("node configuration: MirrorNode ApiAddress is required")
	}
	m.ApiAddress = cfg.ApiAddress

	m.PollingInterval = defaultPollingInterval
	m.QueryMaxLimit = defaultQueryMaxLimit
	m.QueryDefaultLimit = defaultQueryDefaultLimit
	m.RequestTimeout = defaultRequestTimeout

	if cfg.PollingInterval != 0 {
		m.PollingInterval = cfg.PollingInterval
	}
	if cfg.QueryMaxLimit != 0 {
		m.QueryMaxLimit = cfg.QueryMaxLimit
	}
	if cfg.QueryDefaultLimit != 0 {
		m.QueryDefaultLimit = cfg.QueryDefaultLimit
	}
	if cfg.RequestTimeout != 0 {
		m.RequestTimeout = time.Duration(cfg.RequestTimeout)
	}

	m.RetryPolicy = *m.RetryPolicy.DefaultOrConfig(&cfg.RetryPolicy)

	return m
}

type RetryPolicy struct {
	MaxRetry  int
	MinWait   time.Duration
	MaxWait   time.Duration
	MaxJitter time.Duration
}

const (
	// in seconds
	defaultMaxRetry  = 10
	defaultMinWait   = 1
	defaultMaxWait   = 60
	defaultMaxJitter = 0
)

func (r *RetryPolicy) DefaultOrConfig(cfg *parser.RetryPolicy) *RetryPolicy {
	r.MaxRetry = defaultMaxRetry
	r.MinWait = defaultMinWait
	r.MaxWait = defaultMaxWait
	r.MaxJitter = defaultMaxJitter

	if cfg.MaxRetry != 0 {
		r.MaxRetry = cfg.MaxRetry
	}
	if cfg.MinWait != 0 {
		r.MinWait = time.Duration(cfg.MinWait)
	}
	if cfg.MaxWait != 0 {
		r.MaxWait = time.Duration(cfg.MaxWait)
	}
	if cfg.MaxJitter != 0 {
		r.MaxJitter = time.Duration(cfg.MaxJitter)
	}
	return r
}

type Monitoring struct {
	Enable           bool
	DashboardPolling time.Duration
}

type Recovery struct {
	StartTimestamp int64
	StartBlock     int64
}

func New(node parser.Node) Node {
	rpc := parseRpc(node.Clients.Hedera.Rpc)
	config := Node{
		Database: Database(node.Database),
		Clients: Clients{
			Hedera: Hedera{
				Operator:       Operator(node.Clients.Hedera.Operator),
				Network:        node.Clients.Hedera.Network,
				StartTimestamp: node.Clients.Hedera.StartTimestamp,
				Rpc:            rpc,
			},
			MirrorNode: *new(MirrorNode).DefaultOrConfig(&node.Clients.MirrorNode),
			Evm:        make(map[uint64]Evm),
			CoinGecko: CoinGecko{
				ApiAddress: node.Clients.CoinGecko.ApiAddress,
			},
			CoinMarketCap: CoinMarketCap{
				ApiKey:     node.Clients.CoinMarketCap.ApiKey,
				ApiAddress: node.Clients.CoinMarketCap.ApiAddress,
			},
		},
		LogLevel:  node.LogLevel,
		Port:      node.Port,
		Validator: node.Validator,
		Monitoring: Monitoring{
			Enable:           node.Monitoring.Enable,
			DashboardPolling: node.Monitoring.DashboardPolling,
		},
	}

	for key, value := range node.Clients.Evm {
		config.Clients.Evm[key] = Evm(value)
	}

	return config
}

func parseRpc(rpcClients map[string]string) map[string]hedera.AccountID {
	res := make(map[string]hedera.AccountID)
	for key, value := range rpcClients {
		nodeAccountID, err := hedera.AccountIDFromString(value)
		if err != nil {
			log.Fatalf("Hedera RPC [%s] failed to parse Node Account ID [%s]. Error: [%s]", key, value, err)
		}
		res[key] = nodeAccountID
	}
	return res
}
