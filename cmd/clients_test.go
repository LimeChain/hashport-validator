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

package main

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	tc "github.com/limechain/hedera-eth-bridge-validator/test/test-config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepareClients(t *testing.T) {
	clients := PrepareClients(tc.TestConfig.Validator.Clients)
	assert.NotEmpty(t, clients)

	assert.IsType(t, &ethereum.Client{}, clients.Ethereum)
	assert.IsType(t, &hedera.Node{}, clients.HederaNode)
	assert.IsType(t, &mirror_node.Client{}, clients.MirrorNode)

	assert.NotEmpty(t, clients.Ethereum)
	assert.NotEmpty(t, clients.HederaNode)
	assert.NotEmpty(t, clients.MirrorNode)
}
