/*
 * Copyright 2021 LimeChain Ltd.
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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	log "github.com/sirupsen/logrus"
	"time"
)

type Node struct {
	Database  Database
	Clients   Clients
	LogLevel  string
	Port      string
	Validator bool
}

type Database struct {
	Host     string
	Name     string
	Password string
	Port     string
	Username string
}

type Clients struct {
	Evm        map[int64]Evm
	Hedera     Hedera
	MirrorNode MirrorNode
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

type MirrorNode struct {
	ClientAddress   string
	ApiAddress      string
	PollingInterval time.Duration
}

type Recovery struct {
	StartTimestamp int64
	StartBlock     int64
}

func New(node parser.Node) Node {
	rpc := make(map[string]hedera.AccountID)
	for key, value := range node.Clients.Hedera.Rpc {
		nodeAccoundID, err := hedera.AccountIDFromString(value)
		if err != nil {
			log.Fatalf("Hedera RPC [%s] failed to parse Node Account ID [%s]. Error: [%s]", key, value, err)
		}
		rpc[key] = nodeAccoundID
	}

	config := Node{
		Database: Database(node.Database),
		Clients: Clients{
			Hedera: Hedera{
				Operator:       Operator(node.Clients.Hedera.Operator),
				Network:        node.Clients.Hedera.Network,
				StartTimestamp: node.Clients.Hedera.StartTimestamp,
				Rpc:            rpc,
			},
			MirrorNode: MirrorNode(node.Clients.MirrorNode),
			Evm:        make(map[int64]Evm),
		},
		LogLevel:  node.LogLevel,
		Port:      node.Port,
		Validator: node.Validator,
	}

	for key, value := range node.Clients.Evm {
		config.Clients.Evm[key] = Evm(value)
	}

	return config
}
