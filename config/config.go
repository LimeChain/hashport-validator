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
	"github.com/caarlos0/env/v6"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	defaultBridgeFile = "config/bridge.yml"
	defaultNodeFile   = "config/node.yml"
)

func LoadConfig() Config {
	var parsed parser.Config
	GetConfig(&parsed, defaultBridgeFile)
	GetConfig(&parsed, defaultNodeFile)

	if err := env.Parse(&parsed); err != nil {
		panic(err)
	}

	return Config{
		Node:         New(parsed.Node),
		Bridge:       NewBridge(parsed.Bridge),
		ParsedBridge: parsed.Bridge,
	}
}

func GetConfig(config interface{}, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	filename, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

type Config struct {
	Node         Node
	Bridge       Bridge
	ParsedBridge parser.Bridge
}
