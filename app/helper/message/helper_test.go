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

package message

import (
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	model "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	fungibleSignatureMsg = &model.TopicEthSignatureMessage{
		SourceChainId: constants.OldHederaNetworkId,
		TargetChainId: constants.OldHederaNetworkId,
	}

	nonFungibleSignatureMsg = &model.TopicEthNftSignatureMessage{
		SourceChainId: constants.OldHederaNetworkId,
		TargetChainId: constants.OldHederaNetworkId,
	}
)

func Test_UpdateHederaChainIdOfFungibleMsg(t *testing.T) {
	UpdateHederaChainIdOfFungibleMsg(fungibleSignatureMsg)
	assert.Equal(t, fungibleSignatureMsg.SourceChainId, constants.HederaNetworkId)
	assert.Equal(t, fungibleSignatureMsg.TargetChainId, constants.HederaNetworkId)
}

func Test_UpdateHederaChainIdOfNftMsg(t *testing.T) {
	UpdateHederaChainIdOfNftMsg(nonFungibleSignatureMsg)
	assert.Equal(t, nonFungibleSignatureMsg.SourceChainId, constants.HederaNetworkId)
	assert.Equal(t, nonFungibleSignatureMsg.TargetChainId, constants.HederaNetworkId)
}
