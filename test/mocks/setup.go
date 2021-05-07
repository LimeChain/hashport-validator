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

package mocks

import (
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/db"
	hedera_mirror_client "github.com/limechain/hedera-eth-bridge-validator/test/mocks/hedera-mirror-client"
	hedera_node_client "github.com/limechain/hedera-eth-bridge-validator/test/mocks/hedera-node-client"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/rate-provider"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/service"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/transaction"
)

var MExchangeRateProvider *rate_provider.MockExchangeRateProvider
var MTransferService *service.MockTransferService
var MDistributorService *service.MockDistrubutorService
var MScheduledService *service.MockScheduledService
var MFeeService *service.MockFeeService
var MBridgeContractService *MockBridgeContract
var MBurnEventRepository *db.MockBurnEventRepository
var MFeeRepository *db.MockFeeRepository
var MHederaMirrorClient *hedera_mirror_client.MockHederaMirrorClient
var MHederaNodeClient *hedera_node_client.MockHederaNodeClient
var MHederaTransactionResponse *transaction.MockTransactionResponse

func Setup() {
	MBridgeContractService = &MockBridgeContract{}
	MExchangeRateProvider = &rate_provider.MockExchangeRateProvider{}
	MTransferService = &service.MockTransferService{}
	MScheduledService = &service.MockScheduledService{}
	MFeeService = &service.MockFeeService{}
	MBurnEventRepository = &db.MockBurnEventRepository{}
	MFeeRepository = &db.MockFeeRepository{}
	MDistributorService = &service.MockDistrubutorService{}
	MHederaMirrorClient = &hedera_mirror_client.MockHederaMirrorClient{}
	MHederaNodeClient = &hedera_node_client.MockHederaNodeClient{}
	MHederaTransactionResponse = &transaction.MockTransactionResponse{}
}
