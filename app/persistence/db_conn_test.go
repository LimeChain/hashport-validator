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

package persistence

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	connector *pgConnector
	cfg       config.Database
)

func setupConnector() {
	mocks.Setup()

	cfg = config.Database{
		Host:     "host",
		Name:     "name",
		Password: "password",
		Port:     "4200",
		Username: "username",
	}

	connector = NewPgConnector(cfg)
}

func Test_NewPgConnector(t *testing.T) {
	setupConnector()

	actual := NewPgConnector(cfg)
	assert.Equal(t, connector, actual)
}


