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
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/client"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/database"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/handlers"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/http"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/queue"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/rate-provider"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/repository"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/service"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/watchers"
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
var MHederaMirrorClient *client.MockHederaMirror
var MHederaNodeClient *client.MockHederaNode
var MEVMCoreClient *client.MockEVMCore
var MHTTPClient *client.MockHttp
var MDiamondRouter *client.MockDiamondRouter
var MReadOnlyService *service.MockReadOnlyService
var MEVMClient *client.MockEVM
var MEVMTokenClient *client.MockEVMToken
var MPricingClient *client.MockPricingClient
var MSignerService *service.MockSignerService
var MDatabase *database.MockDatabase
var MQueue *queue.MockQueue
var MPrometheusService *service.MockPrometheusService
var MAssetsService *service.MockAssetsService
var MPricingService *service.MockPricingService
var MResponseWriter *http.MockResponseWriter
var MWatcher *watchers.MockWatcher
var MHandler *handlers.MockHandler

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
	MHederaMirrorClient = &client.MockHederaMirror{}
	MHederaNodeClient = &client.MockHederaNode{}
	MEVMClient = &client.MockEVM{}
	MEVMCoreClient = &client.MockEVMCore{}
	MEVMTokenClient = &client.MockEVMToken{}
	MHTTPClient = &client.MockHttp{}
	MDiamondRouter = &client.MockDiamondRouter{}
	MQueue = &queue.MockQueue{}
	MPrometheusService = &service.MockPrometheusService{}
	MAssetsService = &service.MockAssetsService{}
	MPricingService = &service.MockPricingService{}
	MPricingClient = &client.MockPricingClient{}
	MResponseWriter = &http.MockResponseWriter{}
	MWatcher = &watchers.MockWatcher{}
	MHandler = &handlers.MockHandler{}
}
