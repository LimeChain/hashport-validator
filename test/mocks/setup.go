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

package mocks

import (
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/database"
	evm_client "github.com/limechain/hedera-eth-bridge-validator/test/mocks/evm-client"
	evm_token_client "github.com/limechain/hedera-eth-bridge-validator/test/mocks/evm-token-client"
	hedera_mirror_client "github.com/limechain/hedera-eth-bridge-validator/test/mocks/hedera-mirror-client"
	hedera_node_client "github.com/limechain/hedera-eth-bridge-validator/test/mocks/hedera-node-client"
	http_client "github.com/limechain/hedera-eth-bridge-validator/test/mocks/http-client"
	pricing_client "github.com/limechain/hedera-eth-bridge-validator/test/mocks/pricing-client"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/queue"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/rate-provider"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/repository"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/service"
)

var MExchangeRateProvider *rate_provider.MockExchangeRateProvider
var MTransferService *service.MockTransferService
var MDistributorService *service.MockDistrubutorService
var MMessageService *service.MockMessageService
var MScheduledService *service.MockScheduledService
var MFeeService *service.MockFeeService
var MBurnService *service.MockBurnService
var MLockService *service.MockLockService
var MBridgeContractService *MockBridgeContract
var MTransferRepository *repository.MockTransferRepository
var MMessageRepository *repository.MockMessageRepository
var MFeeRepository *repository.MockFeeRepository
var MScheduleRepository *repository.MockScheduleRepository
var MStatusRepository *repository.MockStatusRepository
var MHederaMirrorClient *hedera_mirror_client.MockHederaMirrorClient
var MHederaNodeClient *hedera_node_client.MockHederaNodeClient
var MEVMCoreClient *evm_client.MockEVMCoreClient
var MHTTPClient *http_client.MockHttpClient
var MReadOnlyService *service.MockReadOnlyService
var MEVMClient *evm_client.MockEVMClient
var MEVMTokenClient *evm_token_client.MockEVMTokenClient
var MPricingClient *pricing_client.MockPricingClient
var MSignerService *service.MockSignerService
var MDatabase *database.MockDatabase
var MQueue *queue.MockQueue
var MPrometheusService *service.MockPrometheusService
var MAssetsService *service.MockAssetsService
var MPricingService *service.MockPricingService

func Setup() {
	MDatabase = &database.MockDatabase{}
	MBridgeContractService = &MockBridgeContract{}
	MExchangeRateProvider = &rate_provider.MockExchangeRateProvider{}
	MTransferService = &service.MockTransferService{}
	MScheduledService = &service.MockScheduledService{}
	MFeeService = &service.MockFeeService{}
	MSignerService = &service.MockSignerService{}
	MLockService = &service.MockLockService{}
	MBurnService = &service.MockBurnService{}
	MTransferRepository = &repository.MockTransferRepository{}
	MFeeRepository = &repository.MockFeeRepository{}
	MMessageRepository = &repository.MockMessageRepository{}
	MScheduleRepository = &repository.MockScheduleRepository{}
	MStatusRepository = &repository.MockStatusRepository{}
	MDistributorService = &service.MockDistrubutorService{}
	MReadOnlyService = &service.MockReadOnlyService{}
	MMessageService = &service.MockMessageService{}
	MHederaMirrorClient = &hedera_mirror_client.MockHederaMirrorClient{}
	MHederaNodeClient = &hedera_node_client.MockHederaNodeClient{}
	MEVMClient = &evm_client.MockEVMClient{}
	MEVMCoreClient = &evm_client.MockEVMCoreClient{}
	MEVMTokenClient = &evm_token_client.MockEVMTokenClient{}
	MHTTPClient = &http_client.MockHttpClient{}
	MQueue = &queue.MockQueue{}
	MPrometheusService = &service.MockPrometheusService{}
	MAssetsService = &service.MockAssetsService{}
	MPricingService = &service.MockPricingService{}
	MPricingClient = &pricing_client.MockPricingClient{}

}
