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

package main

import (
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/server"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/recovery"
	"github.com/limechain/hedera-eth-bridge-validator/bootstrap"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Config
	configuration, parsedBridge, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	config.InitLogger(configuration.Node.LogLevel, configuration.Node.LogFormat)

	// Prepare Clients
	clients := bootstrap.PrepareClients(configuration.Node.Clients, configuration.Bridge.EVMs, parsedBridge.Networks)

	// Prepare Node
	server := server.NewServer()

	var services *bootstrap.Services = nil
	conn := persistence.NewPgConnector(configuration.Node.Database)
	db := persistence.NewDatabase(conn)
	db.Migrate()

	// Prepare repositories
	repositories := bootstrap.PrepareRepositories(db)

	// Prepare Services
	var parsedBridgeConfigTopicId hedera.TopicID
	if !parsedBridge.UseLocalConfig {
		var err error
		parsedBridgeConfigTopicId, err = hedera.TopicIDFromString(parsedBridge.ConfigTopicId)
		if err != nil {
			panic(fmt.Sprintf("failed to parse bridge config topic id [%s]. Err: [%s]", parsedBridgeConfigTopicId, err))
		}
	}
	services = bootstrap.PrepareServices(configuration, parsedBridge, clients, *repositories, parsedBridgeConfigTopicId)
	bootstrap.InitializeServerPairs(server, services, repositories, clients, configuration, parsedBridge, parsedBridgeConfigTopicId)

	apiRouter := bootstrap.InitializeAPIRouter(services, parsedBridge)

	executeRecovery(repositories.Fee, repositories.Schedule, clients.MirrorNode)

	// Start
	server.Run(apiRouter.Router, fmt.Sprintf(":%s", configuration.Node.Port))
}

func executeRecovery(feeRepository repository.Fee, scheduleRepository repository.Schedule, client client.MirrorNode) {
	r := recovery.New(feeRepository, scheduleRepository, client)

	r.Execute()
}
